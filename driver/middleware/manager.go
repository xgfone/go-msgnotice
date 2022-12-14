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

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/xgfone/go-msgnotice/driver"
)

// DefaultManager is the default global middleware manager.
var DefaultManager = NewManager(nil)

type driverWrapper struct{ Driver driver.Driver }
type middlewaresWrapper struct{ Middlewares }

// Manager is used to manage a group of the driver middlewares,
// which has implemented the interface driver.Driver.
type Manager struct {
	orig   atomic.Value
	driver atomic.Value

	maps map[string]Middleware
	lock sync.RWMutex
	mdws atomic.Value
}

// NewManager returns a new middleware manager.
func NewManager(defaultDriver driver.Driver) *Manager {
	m := &Manager{maps: make(map[string]Middleware, 8)}
	m.orig.Store(driverWrapper{Driver: defaultDriver})
	m.updateMiddlewares()
	return m
}

// Clone clones itself and returns a new one.
func (m *Manager) Clone() *Manager {
	_m := NewManager(m.GetDriver())
	_m.Use(m.GetMiddlewares()...)
	return _m
}

// Use is a convenient function to add a group of the given middlewares,
// which will panic with an error when the given middleware has been added.
func (m *Manager) Use(mws ...Middleware) {
	for _, mw := range mws {
		if err := m.AddMiddleware(mw); err != nil {
			panic(err)
		}
	}
}

// Cancel is a convenient function to remove the middlewares by the given names.
func (m *Manager) Cancel(names ...string) {
	for _, name := range names {
		m.DelMiddleware(name)
	}
}

// ResetMiddlewares resets the middlewares.
func (m *Manager) ResetMiddlewares(mws ...Middleware) {
	m.lock.Lock()
	for name := range m.maps {
		delete(m.maps, name)
	}
	for _, mw := range mws {
		m.maps[mw.Name()] = mw
	}
	m.updateMiddlewares()
	m.lock.Unlock()
}

// UpsertMiddlewares adds or updates the middlewares.
func (m *Manager) UpsertMiddlewares(mws ...Middleware) {
	m.lock.Lock()
	for _, mw := range mws {
		m.maps[mw.Name()] = mw
	}
	m.updateMiddlewares()
	m.lock.Unlock()
}

// AddMiddleware adds the middleware.
func (m *Manager) AddMiddleware(mw Middleware) (err error) {
	name := mw.Name()
	m.lock.Lock()
	if _, ok := m.maps[name]; ok {
		err = fmt.Errorf("the middleware named '%s' has existed", name)
	} else {
		m.maps[name] = mw
		m.updateMiddlewares()
	}
	m.lock.Unlock()
	return
}

// DelMiddleware removes and returns the middleware by the name.
//
// If the middleware does not exist, do nothing and return nil.
func (m *Manager) DelMiddleware(name string) Middleware {
	m.lock.Lock()
	mw, ok := m.maps[name]
	if ok {
		delete(m.maps, name)
		m.updateMiddlewares()
	}
	m.lock.Unlock()
	return mw
}

// GetMiddleware returns the middleware by the name.
//
// If the middleware does not exist, return nil.
func (m *Manager) GetMiddleware(name string) Middleware {
	m.lock.RLock()
	mw := m.maps[name]
	m.lock.RUnlock()
	return mw
}

// GetMiddlewares returns all the middlewares.
func (m *Manager) GetMiddlewares() Middlewares {
	return m.mdws.Load().(middlewaresWrapper).Middlewares
}

// WrapDriver is equal to m.WrapDriverWithType("", driver).
func (m *Manager) WrapDriver(driver driver.Driver) driver.Driver {
	return m.WrapDriverWithType("", driver)
}

// WrapDriverWithType uses the inner middlewares of the specific type
// to wrap the given driver.
func (m *Manager) WrapDriverWithType(_type string, driver driver.Driver) driver.Driver {
	return m.GetMiddlewares().WrapDriverWithType(_type, driver)
}

func (m *Manager) updateMiddlewares() {
	mdws := make(Middlewares, 0, len(m.maps))
	for _, mw := range m.maps {
		mdws = append(mdws, mw)
	}

	sort.Stable(mdws)
	m.mdws.Store(middlewaresWrapper{mdws})
	m.driver.Store(driverWrapper{mdws.WrapDriver(m.GetDriver())})
}

// SetDriver resets the driver.
func (m *Manager) SetDriver(driver driver.Driver) {
	m.orig.Store(driverWrapper{driver})
	m.driver.Store(driverWrapper{m.WrapDriver(driver)})
}

// GetDriver returns the handler.
func (m *Manager) GetDriver() driver.Driver {
	return m.orig.Load().(driverWrapper).Driver
}

func (m *Manager) getDriver() driver.Driver {
	return m.driver.Load().(driverWrapper).Driver
}

var _ driver.Driver = new(Manager)

// Send implements the interface driver.Driver#Send.
//
// Notice: the driver must be set.
func (m *Manager) Send(c context.Context, msg driver.Message) error {
	return m.getDriver().Send(c, msg)
}

// Stop implements the interface driver.Driver#Stop.
//
// Notice: the driver must be set.
func (m *Manager) Stop() { m.getDriver().Stop() }
