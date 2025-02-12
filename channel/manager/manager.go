// Copyright 2022~2025 xgfone
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
	"github.com/xgfone/go-toolkit/mapx"
)

// Default is the global default channel manager.
var Default = NewManager()

func init() { driver.DefaultSender = Default }

// Manager is used to manage a group of channels.
type Manager struct {
	// WrapDriver is used to wrap a driver and returns a new.
	//
	// Default: middleware.DefaultManager.Driver
	WrapDriver func(driver.Driver) driver.Driver

	// GetChannelName is used to get the channel name by the message
	// when the message name is empty.
	//
	// Default: DefaultMapping.GetFromMessage
	GetChannelName func(context.Context, driver.Message) (string, error)

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

func (m *Manager) getChannelName(ctx context.Context, msg driver.Message) (string, error) {
	if m.GetChannelName != nil {
		return m.GetChannelName(ctx, msg)
	}
	return DefaultMapping.GetFromMessage(ctx, msg)
}

// AddChannel adds the channel with the name.
//
// Notice: it will apply the driver middlewares to the channel driver.
func (m *Manager) AddChannel(ch channel.Channel) (ok bool) {
	m.clock.Lock()
	defer m.clock.Unlock()
	if _, ok = m.chmap[ch.Name]; !ok {
		m.chmap[ch.Name] = channel.New(ch.Name, m.wrap(ch.Driver()))
		m.updateChannels()
	}
	return !ok
}

// UpsertChannels adds or updates a set of channels.
//
// Notice: it will apply the driver middlewares to every channel driver.
func (m *Manager) UpsertChannels(channels ...channel.Channel) {
	if len(channels) == 0 {
		return
	}

	m.clock.Lock()
	defer m.clock.Unlock()
	for _, ch := range channels {
		if oldch, ok := m.chmap[ch.Name]; ok {
			oldch.Driver().Stop()
		}
		m.chmap[ch.Name] = channel.New(ch.Name, m.wrap(ch.Driver()))
	}
	m.updateChannels()
}

// DelChannel deletes the channel by the name.
//
// If the channel does not exist, do nothing.
func (m *Manager) DelChannel(channelName string) {
	m.clock.Lock()
	defer m.clock.Unlock()
	if ch, ok := m.chmap[channelName]; ok {
		ch.Driver().Stop()
		delete(m.chmap, channelName)
		m.updateChannels()
	}
}

// GetChannel returns the channel by the name.
func (m *Manager) GetChannel(channelName string) (channel.Channel, bool) {
	ch, ok := m.getChannels()[channelName]
	return ch, ok
}

// GetChannels returns all the channels.
func (m *Manager) GetChannels() []channel.Channel {
	return mapx.Values(m.getChannels())
}

// Send looks up the channel by the message and send the message with the channel.
func (m *Manager) Send(ctx context.Context, msg driver.Message) (err error) {
	if msg.Name == "" {
		msg.Name, err = m.getChannelName(ctx, msg)
		if err != nil {
			return err
		}
	}

	if ch, ok := m.getChannels()[msg.Name]; ok {
		err = ch.Send(ctx, msg)
	} else {
		err = channel.NotExistError{Channel: msg.Name}
	}

	return
}

// Stop stops all the registered channels.
func (m *Manager) Stop() {
	for _, channel := range m.GetChannels() {
		channel.Driver().Stop()
	}
}
