// Copyright 2024 xgfone
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

package middleware

import (
	"fmt"
	"testing"

	"github.com/xgfone/go-msgnotice/driver"
)

func TestMiddlewaresSort(t *testing.T) {
	noop := func(d driver.Driver) driver.Driver { return d }
	ms := Middlewares{
		New("m1", 1, noop),
		New("m4", 3, noop),
		New("m3", 3, noop),
		New("m2", 2, noop),
	}
	ms.Sort()

	for i, m := range ms {
		name := m.(interface{ Name() string }).Name()
		if expect := fmt.Sprintf("m%d", 4-i); expect != name {
			t.Errorf("expect middleware '%s', but got '%s'", expect, name)
		}
	}
}
