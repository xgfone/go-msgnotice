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

// Package logger provides a driver middleware to log the sent message.
package logger

import (
	"context"
	"log"
	"time"

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/middleware"
)

// Event represents a log message event.
type Event struct {
	driver.Message
	Channel *channel.Channel
	Start   time.Time
	Err     error
}

// LogEvent is used to log the message event.
//
// If not set, use the default by using log.Printf.
var LogEvent func(Event)

func logevent(e Event) {
	if LogEvent != nil {
		LogEvent(e)
	} else {
		var cname, dname string
		if e.Channel != nil {
			cname = e.Channel.ChannelName
			dname = e.Channel.DriverName
		}

		log.Printf("channel=%s, driver=%s, title=%s, content=%s, receivers=%s, metadata=%v, start=%d, cost=%s, err=%v",
			cname, dname, e.Title, e.Content, e.Receiver, e.Metadata, e.Start.Unix(), time.Since(e.Start), e.Err)
	}
}

// New returns a new logger middleware to log the sent message.
func New(_type string, priority int) middleware.Middleware {
	return middleware.NewMiddleware("logger", _type, priority, func(d driver.Driver) driver.Driver {
		return driver.New(func(c context.Context, m driver.Message) error {
			start := time.Now()
			err := d.Send(c, m)
			ch := channel.GetChannelFromContext(c)
			logevent(Event{Channel: ch, Message: m, Start: start, Err: err})
			return err
		}, d.Stop)
	})
}
