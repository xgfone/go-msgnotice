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

package builder

import (
	"fmt"
	"sync"

	"github.com/xgfone/go-msgnotice/driver"
)

var (
	block    = new(sync.RWMutex)
	builders = make(map[string]Builder, 16)
)

// NewAndRegister news the driver builder and register it.
func NewAndRegister(name, _type string, newDriver func(map[string]interface{}) (driver.Driver, error)) {
	Register(New(name, _type, newDriver))
}

// Register registers the driver builder.
//
// If the builder has been registered, override it.
func Register(builder Builder) {
	block.Lock()
	defer block.Unlock()
	builders[builder.Name()] = builder
}

// Unregister unregisters the driver builder by the name.
func Unregister(name string) {
	block.Lock()
	delete(builders, name)
	block.Unlock()
}

// Get returns the driver builder by the name.
//
// Return nil if the builder does not exist.
func Get(name string) Builder {
	block.RLock()
	builder := builders[name]
	block.RUnlock()
	return builder
}

// GetAll returns all the driver builders.
func GetAll() []Builder {
	block.RLock()
	_builders := make([]Builder, 0, len(builders))
	for _, builder := range builders {
		_builders = append(_builders, builder)
	}
	block.RUnlock()
	return _builders
}

// Build looks up the builder by the name and build the driver with the config.
func Build(name string, config map[string]interface{}) (_type string, driver driver.Driver, err error) {
	if builder := Get(name); builder == nil {
		err = fmt.Errorf("no driver builder named '%s'", name)
	} else if driver, err = builder.New(config); err == nil {
		_type = builder.Type()
	}
	return
}
