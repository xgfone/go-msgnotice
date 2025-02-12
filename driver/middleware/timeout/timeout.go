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

// Package timeout provides a driver middleware to set the context timeout.
package timeout

import (
	"context"
	"time"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/middleware"
)

// New returns a new timeout middleware to set the context timeout.
func New(priority int, timeout time.Duration, matcher driver.Matcher) middleware.Middleware {
	if timeout <= 0 {
		panic("the timeout must be a positive")
	}

	return middleware.NewWithMatch("timeout", priority, matcher, func(d driver.Driver) driver.Driver {
		return driver.Wrap(d, func(c context.Context, m driver.Message, d driver.Driver) error {
			c, cancel := context.WithTimeout(c, timeout)
			defer cancel()
			return d.Send(c, m)
		})
	})
}
