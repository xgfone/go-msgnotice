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

// Package channel provides the channel to send the message.
package channel

import (
	"context"
	"errors"
	"fmt"

	"github.com/xgfone/go-msgnotice/driver"
)

// ErrNoChannel is used to represent the error that the channel does not exist.
var ErrNoChannel = errors.New("no channel")

// Send uses driver.DefaultSender to send the message.
func Send(ctx context.Context, msg driver.Message) error {
	return driver.DefaultSender.Send(ctx, msg)
}

// Channel represents a channel to send the message.
type Channel struct {
	Name  string
	Extra any

	driver driver.Driver `json:"-" yaml:"-" sql:"-" xml:"-"`
}

// New returns a new channel.
func New(name string, driver driver.Driver) Channel {
	if driver == nil {
		panic("Channel.New: the driver is nil")
	}
	return Channel{Name: name, driver: driver}
}

// WithExtra returns a new channel with the extra.
func (c Channel) WithExtra(extra any) Channel {
	c.Extra = extra
	return c
}

// Driver returns the driver of the channel.
func (c Channel) Driver() driver.Driver {
	return c.driver
}

// Send implements the interface driver.Sender to send the message.
func (c Channel) Send(ctx context.Context, msg driver.Message) error {
	return c.driver.Send(ctx, msg)
}

// String imports the interface fmt.Stringer.
func (c Channel) String() string {
	if c.Name == "" {
		return fmt.Sprintf("Channel(driver=%s)", c.driver.Type())
	}
	return fmt.Sprintf("Channel(name=%s, driver=%s)", c.Name, c.driver.Type())
}

type (
	// NotExistError represents the error that there is not the given channel.
	NotExistError struct {
		Channel string
	}

	// ExistError represents the error that there is the given channel.
	ExistError struct {
		Channel string
	}
)

func (e NotExistError) Error() string {
	return fmt.Sprintf("no channel named '%s'", e.Channel)
}

func (e ExistError) Error() string {
	return fmt.Sprintf("channel named '%s' has existed", e.Channel)
}
