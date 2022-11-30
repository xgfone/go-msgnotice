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
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/driver/middleware"
)

// Default is the global default channel manager.
var Default = NewManager(middleware.DefaultManager)

func init() { channel.Send = Default.Send }

// NewChannelFunc is a function to new a channel from the config.
type NewChannelFunc func(channelName, driverName string, driverConf map[string]interface{}) (*channel.Channel, error)

// Manager is used to manage a group of channels.
type Manager struct {
	// NewChannel is used to create a new channel.
	//
	// Default: channel.NewChannel
	NewChannel NewChannelFunc

	// DriverMiddlewares is used to manage the middlewares of the channel drivers.
	DriverMiddlewares *middleware.Manager

	// GetDefaultChannelName is used to get the default channel name
	// if no channel is given, which should returns ("", nil) if not found
	// the default channel name.
	//
	// Default: use DefaultChannels.GetDefaultChannelName
	GetDefaultChannelName func(c context.Context, metadata map[string]interface{}) (string, error)

	clock    sync.RWMutex
	chmap    map[string]*channel.Channel // for channels
	channels atomic.Value
}

// NewManager returns a new channel manager.
//
// If the driver middleware manager is nil, new one.
func NewManager(driverMiddlewareManager *middleware.Manager) *Manager {
	if driverMiddlewareManager == nil {
		driverMiddlewareManager = middleware.NewManager(nil)
	}

	m := &Manager{
		chmap:             make(map[string]*channel.Channel, 16),
		NewChannel:        channel.NewChannel,
		DriverMiddlewares: driverMiddlewareManager,
	}

	m.updateChannels()
	return m
}

func (m *Manager) updateChannels() {
	_channels := make(map[string]*channel.Channel, len(m.chmap))
	for name, channel := range m.chmap {
		_channels[name] = channel
	}
	m.channels.Store(_channels)
}

func (m *Manager) getChannels() map[string]*channel.Channel {
	return m.channels.Load().(map[string]*channel.Channel)
}

// AddChannel adds the channel with the name into the global cache.
//
// Notice: it will apply the driver middlewares to the channel driver.
func (m *Manager) AddChannel(channel *channel.Channel) (err error) {
	m.clock.Lock()
	defer m.clock.Unlock()
	if _, ok := m.chmap[channel.ChannelName]; ok {
		err = fmt.Errorf("channel named '%s' has been added", channel.ChannelName)
	} else {
		channel.Driver = m.DriverMiddlewares.WrapDriverWithType(channel.DriverType, channel.Driver)
		m.chmap[channel.ChannelName] = channel
		m.updateChannels()
	}

	return
}

// UpsertChannels adds or updates a set of channels.
//
// Notice: it will apply the driver middlewares to every channel driver.
func (m *Manager) UpsertChannels(channels ...*channel.Channel) {
	m.clock.Lock()
	defer m.clock.Unlock()
	for _, ch := range channels {
		ch.Driver = m.DriverMiddlewares.WrapDriverWithType(ch.DriverType, ch.Driver)
		m.chmap[ch.ChannelName] = ch
	}
	m.updateChannels()
}

// DelChannel deletes the channel by the name from the global cache.
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

// GetChannel returns the channel by the name from the global cache.
//
// Return nil if the channel doest not exist.
func (m *Manager) GetChannel(channelName string) *channel.Channel {
	return m.getChannels()[channelName]
}

// GetAllChannels returns all the channels in the global cache.
func (m *Manager) GetAllChannels() []*channel.Channel {
	channels := m.getChannels()
	_channels := make([]*channel.Channel, 0, len(channels))
	for _, channel := range channels {
		_channels = append(_channels, channel)
	}
	return _channels
}

// BuildAndUpsertChannels builds the channels, and add or upsert them.
func (m *Manager) BuildAndUpsertChannels(channels ...channel.Channel) error {
	_channels := make([]*channel.Channel, len(channels))
	for i, ch := range channels {
		_ch, err := m.NewChannel(ch.ChannelName, ch.DriverName, ch.DriverConf)
		if err != nil {
			return err
		}
		_channels[i] = _ch
	}

	m.UpsertChannels(_channels...)
	return nil
}

// Send is a convenient function to look up the channel by the name
// from the global cache and send the message with the channel.
func (m *Manager) Send(c context.Context, channelName, title, content string,
	metadata map[string]interface{}, receivers ...string) (err error) {
	if channelName == "" {
		if m.GetDefaultChannelName != nil {
			channelName, err = m.GetDefaultChannelName(c, metadata)
		} else {
			channelName, err = DefaultChannels.GetDefaultChannelName(c, metadata)
		}

		if err != nil {
			return err
		}
	}

	if channel, ok := m.getChannels()[channelName]; ok {
		err = channel.Send(c, title, content, metadata, receivers...)
	} else {
		err = fmt.Errorf("no channel named '%s'", channelName)
	}

	return
}

// Stop stops all the registered channels.
func (m *Manager) Stop() {
	for _, channel := range m.GetAllChannels() {
		channel.Stop()
	}
}
