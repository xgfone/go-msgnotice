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

// LogEvent is used to log the message event.
//
// If not set, use the default by using log.Printf.
var LogEvent func(middleware.Event)

func logevent(e middleware.Event) {
	if LogEvent != nil {
		LogEvent(e)
	} else {
		var cname string
		if e.Channel != nil {
			cname = e.Channel.ChannelName
		}

		log.Printf("channel=%s, title=%s, content=%s, metadata=%v, receivers=%v, start=%d, cost=%s, err=%v",
			cname, e.Title, e.Content, e.Metadata, e.Receivers, e.Start.Unix(), time.Since(e.Start), e.Err)
	}
}

// New returns a new logger middleware to log the sent message.
func New(priority int) middleware.Middleware {
	return middleware.NewMiddleware("logger", priority, func(d driver.Driver) driver.Driver {
		return &driverImpl{d}
	})
}

type driverImpl struct{ driver.Driver }

func (d *driverImpl) Send(c context.Context, title, content string,
	metadata map[string]interface{}, tos ...string) (err error) {
	start := time.Now()
	err = d.Driver.Send(c, title, content, metadata, tos...)
	logevent(middleware.Event{
		Channel:   channel.GetChannelFromContext(c),
		Title:     title,
		Content:   content,
		Metadata:  metadata,
		Receivers: tos,
		Start:     start,
		Err:       err,
	})
	return
}
