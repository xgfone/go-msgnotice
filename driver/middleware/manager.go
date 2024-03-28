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

package middleware

import "github.com/xgfone/go-msgnotice/driver"

// DefaultManager is the default global middleware manager.
var DefaultManager = NewManager()

// Manager is used to manage a group of the driver middlewares
type Manager struct {
	mdws Middlewares
}

// NewManager returns a new middleware manager.
func NewManager() *Manager { return new(Manager) }

// Clone clones itself and returns a new one.
func (m *Manager) Clone() *Manager {
	return NewManager().Use(m.mdws...)
}

// Use appends the new middlewares.
func (m *Manager) Use(mws ...Middleware) *Manager {
	m.mdws = m.mdws.Append(mws...)
	return m
}

// Clean cleans all the middlewares.
func (m *Manager) Clean() *Manager {
	m.mdws = nil
	return m
}

// Middlewares returns all the middlewares.
func (m *Manager) Middlewares() Middlewares {
	return m.mdws
}

// Driver implements the interface Middleware.
func (m *Manager) Driver(d driver.Driver) driver.Driver {
	return m.mdws.Driver(d)
}

var _ Middleware = DefaultManager
