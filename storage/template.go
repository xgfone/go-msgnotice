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

package storage

import "context"

// Template represents a message template.
type Template struct {
	Name string
	Tmpl string
	Args []string
}

// TemplateStorage is used to manage the template in the stroage.
type TemplateStorage interface {
	// If the template has exists by the name, return an error.
	AddTemplate(context.Context, Template) error

	// Return nil if the template named name does not exist.
	DelTemplate(c context.Context, name string) error

	// Return ZERO if the template named name does not exist.
	GetTemplate(c context.Context, name string) (Template, error)

	// If any of pageNum and pageSize is less than or equal to 0,
	// return all the templates. pageNum starts with 1.
	GetTemplates(c context.Context, pageNum, pageSize int64) ([]Template, error)
}
