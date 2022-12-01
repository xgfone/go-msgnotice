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

// Package stdout provides a driver to send the message to stdout.
package stdout

import (
	"context"
	"fmt"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/builder"
)

func init() { builder.NewAndRegister("stdout", "stdout", New) }

// New returns a new driver, which outputs the message to stdout.
func New(_ map[string]interface{}) (driver.Driver, error) {
	return driver.Sender(func(c context.Context, m driver.Message) error {
		fmt.Printf("title=%s, content=%s, receiver=%s, metadata=%v\n",
			m.Title, m.Content, m.Receiver, m.Metadata)
		return nil
	}), nil
}
