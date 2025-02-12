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

// Package logger provides a driver middleware to log the sent message.
package logger

import (
	"context"
	"log"
	"time"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/middleware"
)

// New returns a new logger middleware to log the sent message.
func New(priority int, matcher driver.Matcher) middleware.Middleware {
	return middleware.NewWithMatch("logger", priority, matcher, func(d driver.Driver) driver.Driver {
		return driver.Wrap(d, func(c context.Context, m driver.Message, d driver.Driver) (err error) {
			start := time.Now()
			err = d.Send(c, m)
			log.Printf("channel=%s, driver=%s, receiver=%s, content=%s, metadata=%v, cost=%s, err=%v",
				m.Name, m.Type, m.Receiver, m.Content, m.Metadata, time.Since(start), err)
			return
		})
	})
}
