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

package builder

import (
	"fmt"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-toolkit/mapx"
)

var (
	builders = make(map[string]Builder, 16)
)

// NewAndRegister news the driver builder and register it.
func NewAndRegister(name string, build BuilderFunc) {
	Register(New(name, build))
}

// Register registers the driver builder.
//
// If the builder has been registered, override it.
func Register(builder Builder) {
	builders[builder.Name()] = builder
}

// Unregister unregisters the driver builder by the name.
func Unregister(name string) {
	delete(builders, name)
}

// Get returns the driver builder by the name.
//
// Return nil if the builder does not exist.
func Get(name string) Builder {
	return builders[name]
}

// Gets returns all the driver builders.
func Gets() []Builder {
	return mapx.Values(builders)
}

// Build looks up the builder by the name and build the driver with the config.
func Build(name string, config map[string]any) (driver driver.Driver, err error) {
	if builder := Get(name); builder != nil {
		driver, err = builder.Build(config)
	} else {
		err = fmt.Errorf("no driver builder named '%s'", name)
	}
	return
}
