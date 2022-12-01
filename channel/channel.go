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

// Package channel provides the channel to send the message.
package channel

import (
	"context"
	"errors"
	"fmt"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/builder"
)

type ctxkey int8

var channelkey = ctxkey(123)

// SetChannelIntoContext sets the channel into the context
// and returns the new context.
func SetChannelIntoContext(ctx context.Context, channel *Channel) context.Context {
	return context.WithValue(ctx, channelkey, channel)
}

// GetChannelFromContext returns the channel from the context.
//
// Return nil if the context don't has the channel.
func GetChannelFromContext(ctx context.Context) *Channel {
	if v := ctx.Value(channelkey); v != nil {
		return v.(*Channel)
	}
	return nil
}

// ErrNoChannel is used to represent the error that the channel does not exist.
var ErrNoChannel = errors.New("no channel")

// Send is the convenient unified function to send the message.
//
// Notice: it will be set to manager.Default.SendWithChannel if importing
// the package "github.com/xgfone/go-msgnotice/channel/manager",
var Send Sender

// Sender is the function to send the message notice by the channel.
type Sender func(ctx context.Context, channelName string, msg driver.Message) error

// Channel represents a channel to send the message.
type Channel struct {
	ChannelName string
	DriverName  string
	DriverType  string
	DriverConf  map[string]interface{}
	IsDefault   bool

	driver.Driver `json:"-" yaml:"-" sql:"-" xml:"-"`
}

// NewChannel returns a new channel.
func NewChannel(channelName, driverName string, driverConf map[string]interface{}) (*Channel, error) {
	driverType, driver, err := builder.Build(driverName, driverConf)
	if err != nil {
		return nil, err
	}

	return &Channel{
		ChannelName: channelName,
		DriverName:  driverName,
		DriverType:  driverType,
		DriverConf:  driverConf,
		Driver:      driver,
	}, nil
}

// MustNewChannel returns a new channel and panics if there is an error.
func MustNewChannel(channelName, driverName string, driverConf map[string]interface{}) *Channel {
	channel, err := NewChannel(channelName, driverName, driverConf)
	if err != nil {
		panic(err)
	}
	return channel
}

func (c *Channel) String() string {
	if c.ChannelName == "" {
		return fmt.Sprintf("Channel(driver=%s)", c.DriverName)
	}
	return fmt.Sprintf("Channel(name=%s, driver=%s)", c.ChannelName, c.DriverName)
}

var _ driver.Driver = new(Channel)

// Send sends the message to the endpoint by the driver.
//
// Notice: it will put the channel itself into the context ctx
// by SetChannelIntoContext, which can be got out from the context
// by GetChannelFromContext.
func (c *Channel) Send(ctx context.Context, m driver.Message) error {
	return c.Driver.Send(SetChannelIntoContext(ctx, c), m)
}

// NoChannelError represents the error that there is not the given channel.
type NoChannelError string

func (e NoChannelError) Error() string {
	return fmt.Sprintf("no channel named '%s'", string(e))
}
