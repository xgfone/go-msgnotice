// Copyright 2022~2024 xgfone
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

// Package driver provides the message driver interface.
package driver

import (
	"context"
	"errors"
)

// ErrNoDriver is used to represent the error that the driver does not exist.
var ErrNoDriver = errors.New("no driver")

// Message is the message information.
type Message struct {
	Type     string
	Title    string
	Content  string
	Receiver string
	Metadata map[string]any
}

// NewMessage returns a message with the given information.
func NewMessage(mtype, receiver, title, content string, metadata map[string]any) Message {
	return Message{
		Type:     mtype,
		Title:    title,
		Content:  content,
		Receiver: receiver,
		Metadata: metadata,
	}
}

type Sender interface {
	Send(context.Context, Message) error
}

// Driver is used to send the message to the endpoint.
type Driver interface {
	Sender
	Type() string
	Stop()
}

// SendFunc is the driver send function.
type SendFunc func(c context.Context, m Message) error

// Send implements the interface Sender.
func (f SendFunc) Send(c context.Context, m Message) error { return f(c, m) }

// Match reports whether the driver matches the driver type or not.
//
// If dtype is empty, return true always.
func Match(driver Driver, dtype string) bool {
	return dtype == "" || driver.Type() == dtype
}

// New returns a new driver.
//
// If stop is nil, Driver#Stop does nothing.
func New(_type string, send SendFunc, stop func()) Driver {
	if _type == "" {
		panic("dirver.New: the driver type must not be empty")
	}
	if send == nil {
		panic("dirver.New: the driver send function must not be nil")
	}
	if stop == nil {
		stop = donothing
	}
	return &driver{typ: _type, send: send, stop: stop}
}

type driver struct {
	typ  string
	send SendFunc
	stop func()
}

func (d driver) Type() string                            { return d.typ }
func (d driver) Send(c context.Context, m Message) error { return d.send(c, m) }
func (d driver) Stop()                                   { d.stop() }
func donothing()                                         {}

// Wrap wraps the driver and returns a new driver.
func Wrap(d Driver, wrap func(context.Context, Message, Driver) error) Driver {
	if d == nil {
		panic("driver.Wrap: driver must not be nil")
	}
	if wrap == nil {
		panic("driver.Wrap: the wrap function must not be nil")
	}
	return &wrappedDriver{wrap: wrap, Driver: d}
}

// MatchAndWrap is the same as Wrap, but only if matching the given driver type.
func MatchAndWrap(dtype string, d Driver, wrap func(c context.Context, m Message, d Driver) error) Driver {
	if d == nil {
		panic("driver.MatchAndWrap: driver must not be nil")
	}
	if wrap == nil {
		panic("driver.MatchAndWrap: the wrap function must not be nil")
	}
	if !Match(d, dtype) {
		return d
	}
	return &wrappedDriver{wrap: wrap, Driver: d}
}

type wrappedDriver struct {
	wrap func(context.Context, Message, Driver) error
	Driver
}

func (d wrappedDriver) Unwrap() Driver { return d.Driver }

func (d wrappedDriver) Send(c context.Context, m Message) error {
	return d.wrap(c, m, d.Driver)
}
