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

// Package middleware provides the dirver middleware function.
package middleware

import (
	"fmt"
	"slices"

	"github.com/xgfone/go-msgnotice/driver"
)

// Middleware is used to wrap a driver and return a new one.
type Middleware interface {
	Driver(driver.Driver) driver.Driver
}

type middleware struct {
	name   string
	prio   int
	match  driver.Matcher
	driver func(driver.Driver) driver.Driver
}

func (m middleware) Name() string  { return m.name }
func (m middleware) Priority() int { return m.prio }
func (m middleware) Driver(d driver.Driver) driver.Driver {
	if m.match == nil || m.match(d) {
		d = m.driver(d)
	}
	return d
}

func (m middleware) String() string {
	return fmt.Sprintf("Middleware(name=%s, priority=%d)", m.name, m.prio)
}

// New returns a new driver middleware with the name and priority.
func New(name string, prio int, f func(driver.Driver) driver.Driver) Middleware {
	return middleware{name: name, prio: prio, driver: f}
}

// NewWithMatch returns a new driver middleware with the name, priority and driver matcher.
func NewWithMatch(name string, prio int, matcher driver.Matcher, f func(driver.Driver) driver.Driver) Middleware {
	return middleware{name: name, prio: prio, match: matcher, driver: f}
}

// Sort sorts a set of middlewares by the priority from high to low.
func Sort(middlewares []Middleware) {
	slices.SortStableFunc(middlewares, func(a, b Middleware) int {
		return getprio(b) - getprio(a)
	})
}

func getprio(m Middleware) int {
	if p, ok := m.(interface{ Priority() int }); ok {
		return p.Priority()
	}
	return 0
}

/// ----------------------------------------------------------------------- ///

var _ Middleware = Middlewares(nil)

type Middlewares []Middleware

// Sort sorts itself by the priority from high to low.
func (ms Middlewares) Sort() { Sort(ms) }

// Clone clones itself and returns a new one.
func (ms Middlewares) Clone() Middlewares {
	return append(Middlewares(nil), ms...)
}

// Append appends the new middlewares.
func (ms Middlewares) Append(news ...Middleware) Middlewares {
	return append(ms, news...)
}

// Driver implements the interface Middleware.
func (ms Middlewares) Driver(d driver.Driver) driver.Driver {
	if d == nil {
		return d
	}

	for _len := len(ms) - 1; _len >= 0; _len-- {
		d = ms[_len].Driver(d)
	}
	return d
}
