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

package file

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/xgfone/go-msgnotice/storage"
	"github.com/xgfone/go-msgnotice/storage/static"
)

// NewTemplateStorage returns a new template storage, which reads all the templates
// from the given file based on the JSON format, for example,
//
//	[
//	    {
//	        "Name": "xxx",
//	        "Tmpl": "the template {arg1} content with the argument {arg2}",
//	        "Args": ["arg1", "arg2"]
//	    }
//	]
func NewTemplateStorage(filepath string) (storage.TemplateStorage, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	} else if len(data) == 0 {
		return static.NewTemplateStorage(), nil
	}

	var templates []storage.Template
	if err = json.Unmarshal(data, &templates); err != nil {
		return nil, err
	}

	for _, template := range templates {
		if template.Name == "" {
			return nil, errors.New("the template name must not be empty")
		}
	}

	return static.NewTemplateStorage(templates...), nil
}
