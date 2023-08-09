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

// Package template provides a middleware to render the message content
// from a template.
package template

import (
	"context"
	"fmt"
	"strings"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/middleware"
)

// Error is used to represents the template error.
type Error struct {
	TmplName string
	error
}

// Template represents a message template.
type Template struct {
	Name string
	Tmpl string
	Args []string
}

// Getter is used to get the template by the name.
type Getter interface {
	GetTemplate(c context.Context, name string) (Template, bool, error)
}

// GetterFunc is the template getter function.
type GetterFunc func(c context.Context, name string) (Template, bool, error)

// GetTemplate implements the interface Getter.
func (f GetterFunc) GetTemplate(c context.Context, name string) (Template, bool, error) {
	return f(c, name)
}

// New returns a new template middleware to render the content
// from a given template with the arguments.
func New(_type string, priority int, getter Getter) middleware.Middleware {
	return middleware.NewMiddleware("template", _type, priority, func(d driver.Driver) driver.Driver {
		return &driverImpl{Driver: d, Getter: getter}
	})
}

type driverImpl struct {
	driver.Driver
	Getter
}

func (d *driverImpl) Send(c context.Context, m driver.Message) (err error) {
	if strings.HasPrefix(m.Content, "tmpl:") {
		m.Content, err = d.render(c, m.Content[len("tmpl:"):], m.Metadata)
		if err != nil {
			return
		}
	}
	return d.Driver.Send(c, m)
}

func (d *driverImpl) render(c context.Context, name string,
	metadata map[string]interface{}) (content string, err error) {
	if name == "" {
		return "", Error{}
	}

	tmpl, ok, err := d.GetTemplate(c, name)
	if err != nil {
		return "", Error{TmplName: name, error: err}
	} else if !ok {
		return "", Error{TmplName: name}
	}

	content = tmpl.Tmpl
	for _, arg := range tmpl.Args {
		placeholder := fmt.Sprintf("{%s}", arg)
		if v, ok := metadata[arg]; ok {
			content = strings.Replace(content, placeholder, fmt.Sprint(v), 1)
		} else {
			content = strings.Replace(content, placeholder, "", 1)
		}
	}

	return
}
