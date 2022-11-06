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
	"time"

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/driver"
)

// Event represents a message event.
type Event struct {
	Channel   *channel.Channel
	Title     string
	Content   string
	Metadata  map[string]interface{}
	Receivers []string
	Start     time.Time
	Err       error
}

// Middleware is the common driver middleware.
type Middleware interface {
	// Name returns the name of the middleware.
	Name() string

	// The smaller the value, the higher the priority and the middleware
	// is executed preferentially.
	Priority() int

	// WrapDriver is used to wrap the driver and returns a new one.
	WrapDriver(wrappedDriver driver.Driver) (newDriver driver.Driver)
}

type middleware struct {
	name     string
	priority int
	driver   func(driver.Driver) driver.Driver
}

func (m middleware) Name() string                             { return m.name }
func (m middleware) Priority() int                            { return m.priority }
func (m middleware) WrapDriver(d driver.Driver) driver.Driver { return m.driver(d) }

// NewMiddleware returns a new common driver middleware.
func NewMiddleware(name string, prio int, f func(driver.Driver) driver.Driver) Middleware {
	return middleware{name: name, priority: prio, driver: f}
}

// Middlewares is a group of the common driver middlewares.
type Middlewares []Middleware

func (ms Middlewares) Len() int           { return len(ms) }
func (ms Middlewares) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms Middlewares) Less(i, j int) bool { return ms[i].Priority() < ms[j].Priority() }

// WrapDriver wraps the driver with the middlewares and returns a new one.
func (ms Middlewares) WrapDriver(driver driver.Driver) driver.Driver {
	if driver == nil {
		return nil
	}

	for _len := len(ms) - 1; _len >= 0; _len-- {
		driver = ms[_len].WrapDriver(driver)
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
