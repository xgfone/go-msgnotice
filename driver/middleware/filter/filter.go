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

// Package filter provides a driver middleware to stop to send the message.
package filter

import (
	"context"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/middleware"
)

// Filter is used to check and filter the message.
type Filter func(context.Context, driver.Message) (filter bool, err error)

// New returns a new middleware to stop to send the message
// if the filter returns true or an error.
func New(_type string, priority int, filter Filter) middleware.Middleware {
	if filter == nil {
		panic("the filter must not be nil")
	}

	return middleware.NewMiddleware("filter", _type, priority, func(d driver.Driver) driver.Driver {
		return driver.NewDriver(func(c context.Context, m driver.Message) error {
			if filter, err := filter(c, m); err != nil || filter {
				return err
			}
			return d.Send(c, m)
		}, d.Stop)
	})
}
