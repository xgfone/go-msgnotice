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

package static

import (
	"context"
	"errors"
	"fmt"

	"github.com/xgfone/go-msgnotice/storage"
)

// NewTemplateStorage returns a new template storage.
func NewTemplateStorage(templates ...storage.Template) storage.TemplateStorage {
	s := tmplStorage{templates: make([]storage.Template, 0, len(templates))}
	for _, template := range templates {
		if template.Name == "" {
			panic("the template name must not be empty")
		}
		if template.Tmpl == "" {
			panic("the template content must not be empty")
		}

		if s.Index(template.Name) < 0 {
			s.templates = append(s.templates, template)
		}
	}
	return s
}

type tmplStorage struct {
	templates []storage.Template
}

func (s tmplStorage) Index(name string) (index int) {
	for i, _len := 0, len(s.templates); i < _len; i++ {
		if s.templates[i].Name == name {
			return i
		}
	}
	return -1
}

func (s tmplStorage) AddTemplate(context.Context, storage.Template) error {
	return errors.New("TemplateFileStorage.AddTemplate is not implemented")
}

func (s tmplStorage) DelTemplate(_ context.Context, name string) error {
	return errors.New("TemplateFileStorage.DelTemplate is not implemented")
}

func (s tmplStorage) GetTemplate(_ context.Context, name string) (storage.Template, error) {
	if index := s.Index(name); index > -1 {
		return s.templates[index], nil
	}
	return storage.Template{}, fmt.Errorf("no template named '%s'", name)
}

func (s tmplStorage) GetTemplates(_ context.Context, pageNum int64, pageSize int64) ([]storage.Template, error) {
	total := int64(len(s.templates))
	start := int64(0)
	end := total

	if pageNum > 0 && pageSize > 0 {
		start = (pageNum - 1) * pageSize
		if total <= start {
			return nil, nil
		}

		end = pageNum * pageSize
		if total < end {
			end = total
		}
	}

	templates := make([]storage.Template, end-start)
	copy(templates, s.templates[start:end])
	return templates, nil
}
