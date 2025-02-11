// Copyright 2022 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package manager is used to manage a set of the message channels.
package manager

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/middleware"
)

// Default is the global default channel manager.
var Default = NewManager()

func init() { channel.Send = Default.SendWithChannel }

// NewChannelFunc is a function to new a channel from the config.
type NewChannelFunc func(channelName, driverName string, driverConf map[string]any) (*channel.Channel, error)

// Manager is used to manage a group of channels.
type Manager struct {
	// WrapDriver is used to wrap a driver and returns a new.
	//
	// Default: middleware.DefaultManager.Driver
	WrapDriver func(driver.Driver) driver.Driver

	// GetDefaultChannelName is used to get the default channel name
	// if no channel is given, which should returns ("", nil) if not found
	// the default channel name.
	//
	// Default: use DefaultChannels.GetDefaultChannelName
	GetDefaultChannelName func(context.Context, driver.Message) (string, error)

	clock    sync.RWMutex
	chmap    map[string]channel.Channel // for channels
	channels atomic.Value
}

// NewManager returns a new channel manager.
func NewManager() *Manager {
	m := &Manager{chmap: make(map[string]channel.Channel, 16)}
	m.updateChannels()
	return m
}

func (m *Manager) updateChannels() {
	_channels := make(map[string]channel.Channel, len(m.chmap))
	for name, channel := range m.chmap {
		_channels[name] = channel
	}
	m.channels.Store(_channels)
}

func (m *Manager) getChannels() map[string]channel.Channel {
	return m.channels.Load().(map[string]channel.Channel)
}

func (m *Manager) wrap(d driver.Driver) driver.Driver {
	if m.WrapDriver != nil {
		return m.WrapDriver(d)
	}
	return middleware.DefaultManager.Driver(d)
}

func (m *Manager) getDefaultName(ctx context.Context, msg driver.Message) (string, error) {
	if m.GetDefaultChannelName != nil {
		return m.GetDefaultChannelName(ctx, msg)
	}
	return DefaultChannels.GetDefaultChannelName(ctx, msg)
}

// AddChannel adds the channel with the name.
//
// Notice: it will apply the driver middlewares to the channel driver.
func (m *Manager) AddChannel(ch channel.Channel) (err error) {
	m.clock.Lock()
	defer m.clock.Unlock()
	if _, ok := m.chmap[ch.ChannelName]; ok {
		err = channel.ExistError{Channel: ch.ChannelName}
	} else {
		ch.Driver = m.wrap(ch.Driver)
		m.chmap[ch.ChannelName] = ch
		m.updateChannels()
	}

	return
}

// UpsertChannels adds or updates a set of channels.
//
// Notice: it will apply the driver middlewares to every channel driver.
func (m *Manager) UpsertChannels(channels ...channel.Channel) {
	m.clock.Lock()
	defer m.clock.Unlock()
	for _, ch := range channels {
		ch.Driver = m.wrap(ch.Driver)
		m.chmap[ch.ChannelName] = ch
	}
	m.updateChannels()
}

// DelChannel deletes the channel by the name.
//
// If the channel does not exist, do nothing.
func (m *Manager) DelChannel(channelName string) {
	m.clock.Lock()
	if _, ok := m.chmap[channelName]; ok {
		delete(m.chmap, channelName)
		m.updateChannels()
	}
	m.clock.Unlock()
}

// GetChannel returns the channel by the name.
func (m *Manager) GetChannel(channelName string) (channel.Channel, bool) {
	ch, ok := m.getChannels()[channelName]
	return ch, ok
}

// GetChannels returns all the channels.
func (m *Manager) GetChannels() []channel.Channel {
	channels := m.getChannels()
	_channels := make([]channel.Channel, 0, len(channels))
	for _, channel := range channels {
		_channels = append(_channels, channel)
	}
	return _channels
}

// SendWithChannel is a convenient function to look up the channel by the name
// and send the message with the channel.
//
// If channelName is empty, use GetDefaultChannelName to get the default channel.
// If there is not the channel, return channel.NoChannelError(channelName).
func (m *Manager) SendWithChannel(ctx context.Context, channelName string, msg driver.Message) (err error) {
	if channelName == "" {
		channelName, err = m.getDefaultName(ctx, msg)
		if err != nil {
			return err
		}
	}

	if ch, ok := m.getChannels()[channelName]; ok {
		channel.SetChannelIntoContext(ctx, ch)
		err = ch.Send(ctx, msg)
	} else {
		err = channel.NotExistError{Channel: channelName}
	}

	return
}

// Send implements the interface driver.Driver#Send,
// which is equal to m.SendWithChannel(ctx, "", msg).
func (m *Manager) Send(ctx context.Context, msg driver.Message) (err error) {
	return m.SendWithChannel(ctx, "", msg)
}

// Stop implements the interface driver.Driver#Stop,
// which stops all the registered channels.
func (m *Manager) Stop() {
	for _, channel := range m.GetChannels() {
		channel.Stop()
	}
}
