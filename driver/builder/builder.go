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

// Package builder provides the builder to build the message driver.
package builder

import "github.com/xgfone/go-msgnotice/driver"

// Builder is used to build a driver with the config.
type Builder interface {
	Name() string
	Build(config map[string]any) (driver.Driver, error)
}

// BuilderFunc is a function to new a driver.
type BuilderFunc func(name string, config map[string]any) (driver.Driver, error)

// New returns a new driver builder.
func New(name string, build BuilderFunc) Builder {
	return builder{name: name, build: build}
}

type builder struct {
	name  string
	build BuilderFunc
}

func (b builder) Name() string                                  { return b.name }
func (b builder) Build(c map[string]any) (driver.Driver, error) { return b.build(b.name, c) }
