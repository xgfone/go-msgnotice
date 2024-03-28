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

package manager

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-msgnotice/driver"
)

// DefaultChannels is the global default channels.
var DefaultChannels = NewMapping()

// Mapping is used to manage the mapping from the driver type to the channel name.
type Mapping struct {
	lock sync.Mutex
	t2nm map[string]string
	t2ns atomic.Value
}

// NewMapping returns a new default channel manager.
func NewMapping() *Mapping {
	m := &Mapping{t2nm: make(map[string]string, 8)}
	m.updateMapping()
	return m
}

func (m *Mapping) updateMapping() {
	t2nm := make(map[string]string, len(m.t2nm))
	for driverType, channelName := range m.t2nm {
		t2nm[driverType] = channelName
	}
	m.t2ns.Store(t2nm)
}

func (m *Mapping) getMapping() map[string]string {
	return m.t2ns.Load().(map[string]string)
}

func (m *Mapping) set(driverType, channelName string) (changed bool) {
	if name, ok := m.t2nm[driverType]; !ok || name != channelName {
		m.t2nm[driverType] = channelName
		changed = true
	}
	return
}

func (m *Mapping) unset(driverType string) (ok bool) {
	if _, ok = m.t2nm[driverType]; ok {
		delete(m.t2nm, driverType)
	}
	return
}

// Set sets the mapping from the driver type to the channel name.
func (m *Mapping) Set(driverType, channelName string) {
	m.lock.Lock()
	if m.set(driverType, channelName) {
		m.updateMapping()
	}
	m.lock.Unlock()
}

// Sets sets a set of the mappings from the driver type to the channel name.
func (m *Mapping) Sets(driverType2channelNames map[string]string) {
	var changed bool
	m.lock.Lock()
	for driverType, channelName := range driverType2channelNames {
		if m.set(driverType, channelName) {
			changed = true
		}
	}

	if changed {
		m.updateMapping()
	}
	m.lock.Unlock()
}

// Unset unsets the mapping from the driver type to the channel name.
func (m *Mapping) Unset(driverType string) {
	m.lock.Lock()
	if m.unset(driverType) {
		m.updateMapping()
	}
	m.lock.Unlock()
}

// Unsets unsets a set of the mappings from the driver type to the channel name.
func (m *Mapping) Unsets(driverTypes ...string) {
	var changed bool
	m.lock.Lock()
	for _, driverType := range driverTypes {
		if m.unset(driverType) {
			changed = true
		}
	}

	if changed {
		m.updateMapping()
	}
	m.lock.Unlock()
}

// Get returns the channel name mapped by the driver type.
func (m *Mapping) Get(driverType string) string {
	return m.getMapping()[driverType]
}

// GetAll returns all the mappings from the driver type to the channel name.
func (m *Mapping) GetAll() (driverType2ChannelNames map[string]string) {
	mappings := m.getMapping()
	driverType2ChannelNames = make(map[string]string, len(mappings))
	for k, v := range mappings {
		driverType2ChannelNames[k] = v
	}
	return
}

// GetDefaultChannelName returns the channel name mapped by the driver type.
//
// NOTICE: the message type is used as the driver type if not empty.
func (m *Mapping) GetDefaultChannelName(c context.Context, dm driver.Message) (string, error) {
	if len(dm.Type) > 0 {
		return m.Get(dm.Type), nil
	}
	return "", nil
}
