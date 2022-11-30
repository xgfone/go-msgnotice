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

// Package middleware provides the dirver middleware function.
package middleware

import (
	"github.com/xgfone/go-msgnotice/driver"
)

// Middleware is the common driver middleware.
type Middleware interface {
	// Name returns the name of the middleware.
	Name() string

	// Type returns the type of the middleware, which only apply the driver
	// of the specific type.
	//
	// If empty, represents the common middleware and maybe apply any drivers.
	Type() string

	// The smaller the value, the higher the priority and the middleware
	// is executed preferentially.
	Priority() int

	// WrapDriver is used to wrap the driver and returns a new one.
	WrapDriver(wrappedDriver driver.Driver) (newDriver driver.Driver)
}

type middleware struct {
	name     string
	_type    string
	priority int
	driver   func(driver.Driver) driver.Driver
}

func (m middleware) Name() string                             { return m.name }
func (m middleware) Type() string                             { return m._type }
func (m middleware) Priority() int                            { return m.priority }
func (m middleware) WrapDriver(d driver.Driver) driver.Driver { return m.driver(d) }

// NewMiddleware returns a new common driver middleware.
func NewMiddleware(name, _type string, prio int, f func(driver.Driver) driver.Driver) Middleware {
	return middleware{name: name, _type: _type, priority: prio, driver: f}
}

// Middlewares is a group of the common driver middlewares.
type Middlewares []Middleware

func (ms Middlewares) Len() int           { return len(ms) }
func (ms Middlewares) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms Middlewares) Less(i, j int) bool { return ms[i].Priority() < ms[j].Priority() }

// WrapDriver is equal to ms.WrapDriverWithType("", driver).
func (ms Middlewares) WrapDriver(driver driver.Driver) driver.Driver {
	return ms.WrapDriverWithType("", driver)
}

// WrapDriverWithType wraps the driver with the middlewares of the specific type
// and returns a new one.
//
// if _type is empty, apply all the middlewares to the given driver.
func (ms Middlewares) WrapDriverWithType(_type string, driver driver.Driver) driver.Driver {
	if driver == nil {
		return nil
	}

	for _len := len(ms) - 1; _len >= 0; _len-- {
		if mt := ms[_len].Type(); mt == "" || _type == "" || mt == _type {
			driver = ms[_len].WrapDriver(driver)
		}
	}
	return driver
}

// Clone clones itself and appends the new middlewares to the new.
func (ms Middlewares) Clone(news ...Middleware) Middlewares {
	mws := make(Middlewares, len(ms)+len(news))
	copy(mws, ms)
	copy(mws[len(ms):], news)
	return mws
}
