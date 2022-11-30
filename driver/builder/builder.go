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

// Package builder provides the builder to build the message driver.
package builder

import "github.com/xgfone/go-msgnotice/driver"

// Builder is used to build a driver with the config.
type Builder interface {
	Type() string
	Name() string
	New(config map[string]interface{}) (driver.Driver, error)
}

// NewDriverFunc is a function to new a driver.
type NewDriverFunc func(map[string]interface{}) (driver.Driver, error)

// New returns a new driver builder.
func New(name, _type string, newDriver NewDriverFunc) Builder {
	return builder{name: name, typ: _type, new: newDriver}
}

type builder struct {
	name string
	typ  string
	new  NewDriverFunc
}

func (b builder) Type() string                                        { return b.typ }
func (b builder) Name() string                                        { return b.name }
func (b builder) New(c map[string]interface{}) (driver.Driver, error) { return b.new(c) }
