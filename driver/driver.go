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

// Package driver provides the message driver interface.
package driver

import (
	"context"
	"errors"
	"slices"
)

// ErrNoDriver is used to represent the error that the driver does not exist.
var ErrNoDriver = errors.New("no driver")

// DefaultSender is the default sender to send the message.
var DefaultSender Sender

// Message is the message information.
type Message struct {
	Name string // The message or channel name
	Type string // The message or driver type

	Content  any
	Receiver string
	Metadata map[string]any
}

// NewMessage returns a message.
func NewMessage(cname, dtype, receiver string, content any, metadata map[string]any) Message {
	return Message{
		Name: cname,
		Type: dtype,

		Content:  content,
		Receiver: receiver,
		Metadata: metadata,
	}
}

// Sender is the interface to send the message.
type Sender interface {
	Send(context.Context, Message) error
}

// Driver is used to send the message to the endpoint.
type Driver interface {
	Name() string
	Type() string

	Sender
	Stop()
}

// SenderFunc is the message send function.
type SenderFunc func(c context.Context, m Message) error

// Send implements the interface Sender.
func (f SenderFunc) Send(c context.Context, m Message) error { return f(c, m) }

// Matcher is used to check whether to matche the driver or not.
type Matcher func(Driver) bool

// TypesMatcher returns a driver matcher to report whether the driver is matched against the given types.
func TypesMatcher(types ...string) Matcher {
	return func(d Driver) bool {
		return slices.Contains(types, d.Type())
	}
}

// New returns a new common driver.
//
// If stop is nil, `Driver.Stop` does nothing.
func New(name, dtype string, send SenderFunc, stop func()) Driver {
	if name == "" {
		panic("dirver.New: the driver name must not be empty")
	}
	if dtype == "" {
		panic("dirver.New: the driver type must not be empty")
	}
	if send == nil {
		panic("dirver.New: the driver send function must not be nil")
	}
	if stop == nil {
		stop = donothing
	}
	return &driver{typ: dtype, name: name, send: send, stop: stop}
}

type driver struct {
	typ  string
	name string
	send SenderFunc
	stop func()
}

func (d driver) Type() string                            { return d.typ }
func (d driver) Name() string                            { return d.name }
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

type wrappedDriver struct {
	wrap func(context.Context, Message, Driver) error
	Driver
}

func (d wrappedDriver) Unwrap() Driver { return d.Driver }

func (d wrappedDriver) Send(c context.Context, m Message) error {
	return d.wrap(c, m, d.Driver)
}
