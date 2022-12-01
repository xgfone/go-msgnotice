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

// Package validator provides a driver middleware to validate
// whether the receiver is valid.
package validator

import (
	"context"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/middleware"
)

// New returns a new driver middleware to validate whether the receiver is valid.
func New(_type string, priority int, validate func(receiver string) error) middleware.Middleware {
	if validate == nil {
		panic("the validate must not be nil")
	}

	return middleware.NewMiddleware("validator", _type, priority, func(d driver.Driver) driver.Driver {
		return driver.NewDriver(func(c context.Context, m driver.Message) (err error) {
			if err = validate(m.Receiver); err == nil {
				err = d.Send(c, m)
			}
			return
		}, d.Stop)
	})
}
