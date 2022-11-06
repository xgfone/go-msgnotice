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

func setChannelIntoContext(ctx context.Context, channel *Channel) context.Context {
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
var Send func(c context.Context, channelName, title, content string, metadata map[string]interface{}, tos ...string) error

// Channel represents a channel to send the message.
type Channel struct {
	ChannelName string
	DriverName  string
	DriverType  string
	DriverConf  map[string]interface{}

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

func (c *Channel) String() string {
	if c.ChannelName == "" {
		return fmt.Sprintf("Channel(driver=%s)", c.DriverName)
	}
	return fmt.Sprintf("Channel(name=%s, driver=%s)", c.ChannelName, c.DriverName)
}

var _ driver.Driver = new(Channel)

// Send sends the message to the endpoint by the driver.
//
// Notice: it will put the channel itself into the context ctx,
// which can be got out from the context by calling GetChannelFromContext.
func (c *Channel) Send(ctx context.Context, title, content string, md map[string]interface{}, tos ...string) error {
	return c.Driver.Send(setChannelIntoContext(ctx, c), title, content, md, tos...)
}
