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

// Package driver provides the message driver interface.
package driver

import (
	"context"
	"errors"
)

// ErrNoDriver is used to represent the error that the driver does not exist.
var ErrNoDriver = errors.New("no driver")

// Driver is used to send the message to the endpoint.
type Driver interface {
	Send(c context.Context, title, content string, metadata map[string]interface{}, tos ...string) error
	Stop()
}

// Sender is the driver send function.
type Sender func(c context.Context, title, content string, md map[string]interface{}, tos ...string) error

// Send implements the interface Driver#Send.
func (s Sender) Send(c context.Context, title, content string, md map[string]interface{}, tos ...string) error {
	return s(c, title, content, md, tos...)
}

// Stop implements the interface Driver#Stop, which does nothing.
func (s Sender) Stop() {}

// NewDriver returns a new driver from the send and stop functions.
//
// If stop is nil, Driver#Stop does nothing.
func NewDriver(send Sender, stop func()) Driver {
	if send == nil {
		panic("the driver send function must not be nil")
	}
	return driver{send: send, stop: stop}
}

type driver struct {
	send Sender
	stop func()
}

func (d driver) Stop() {
	if d.stop != nil {
		d.stop()
	}
}

func (d driver) Send(c context.Context, t, cnt string, md map[string]interface{}, tos ...string) error {
	return d.send(c, t, cnt, md, tos...)
}
