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
)

// Default is the global default channel manager.
var Default = NewManager()

func init() { channel.Send = Default.Send }

// Manager is used to manage a group of channels.
type Manager struct {
	// NewChannel is used to create a new channel.
	//
	// Default: channel.NewChannel
	NewChannel func(channelName, driverName string, driverConf map[string]interface{}) (*channel.Channel, error)

	clock sync.RWMutex
	chmap map[string]*channel.Channel

	channels atomic.Value
	_default atomic.Value
}

// NewManager returns a new channel manager.
func NewManager() *Manager {
	m := &Manager{
		chmap:      make(map[string]*channel.Channel, 16),
		NewChannel: channel.NewChannel,
	}

	m.SetDefaultChannelName("")
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
func (m *Manager) AddChannel(channel *channel.Channel) (err error) {
	m.clock.Lock()
	defer m.clock.Unlock()
	if _, ok := m.chmap[channel.ChannelName]; ok {
		err = fmt.Errorf("channel named '%s' has been added", channel.ChannelName)
	} else {
		m.chmap[channel.ChannelName] = channel
		m.updateChannels()
	}

	return
}

// UpsertChannel adds or updates the channel.
func (m *Manager) UpsertChannel(channel *channel.Channel) {
	m.clock.Lock()
	defer m.clock.Unlock()
	m.chmap[channel.ChannelName] = channel
	m.updateChannels()
}

// UpsertChannels adds or updates a set of channels.
func (m *Manager) UpsertChannels(channels ...*channel.Channel) {
	m.clock.Lock()
	defer m.clock.Unlock()
	for _, channel := range channels {
		m.chmap[channel.ChannelName] = channel
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

// GetChannels returns all the channels in the global cache.
func (m *Manager) GetChannels() []*channel.Channel {
	channels := m.getChannels()
	_channels := make([]*channel.Channel, 0, len(channels))
	for _, channel := range channels {
		_channels = append(_channels, channel)
	}
	return _channels
}

// SetDefaultChannelName resets the default channel name.
func (m *Manager) SetDefaultChannelName(name string) { m._default.Store(name) }

// GetDefaultChannelName returns the default channel name.
func (m *Manager) GetDefaultChannelName() string { return m._default.Load().(string) }

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
func (m *Manager) Send(c context.Context, channelName, title, content string, metadata map[string]interface{}, tos ...string) error {
	if channelName == "" {
		channelName = m.GetDefaultChannelName()
	}

	if channel, ok := m.getChannels()[channelName]; ok {
		return channel.Send(c, title, content, metadata, tos...)
	}
	return fmt.Errorf("no channel named '%s'", channelName)
}

// Stop stops all the registered channels.
func (m *Manager) Stop() {
	for _, channel := range m.GetChannels() {
		channel.Stop()
	}
}
