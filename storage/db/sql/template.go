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

package dbsql

import (
	"context"
	"strings"
	"time"

	"github.com/xgfone/go-msgnotice/storage"
	"github.com/xgfone/go-sqlx"
)

// NewTemplateStorage returns a new template storage based on database/sql.
func NewTemplateStorage(db *sqlx.DB) storage.TemplateStorage {
	return tmplStorage{Table: sqlx.NewTable("template").WithDB(db)}
}

type templateM struct {
	sqlx.Base

	Name string `sql:"name"`
	Tmpl string `sql:"tmpl"`
	Args string `sql:"args"`
}

type tmplStorage struct{ sqlx.Table }

func (s tmplStorage) AddTemplate(c context.Context, tmpl storage.Template) error {
	_, err := s.InsertInto().Struct(templateM{
		Name: tmpl.Name,
		Tmpl: tmpl.Tmpl,
		Args: strings.Join(tmpl.Args, ","),
	}).ExecContext(c)
	return err
}

func (s tmplStorage) DelTemplate(c context.Context, name string) error {
	_, err := s.Update(sqlx.ColumnDeletedAt.Set(time.Now())).Where(
		sqlx.ColumnDeletedAt.Eq(sqlx.DateTimeZero),
		sqlx.NewColumn("name").Eq(name),
	).ExecContext(c)
	return err
}

func (s tmplStorage) GetTemplate(c context.Context, name string) (storage.Template, error) {
	var tm templateM
	err := s.SelectStruct(c).Where(
		sqlx.ColumnDeletedAt.Eq(sqlx.DateTimeZero),
		sqlx.NewColumn("name").Eq(name),
	).BindRowStructContext(c, &tm)

	if exist, err := sqlx.CheckErrNoRows(err); err != nil || !exist {
		return storage.Template{}, err
	}

	tmpl := storage.Template{Name: tm.Name, Tmpl: tm.Tmpl}
	if tm.Args != "" {
		tmpl.Args = strings.Split(tm.Args, ",")
	}

	return tmpl, nil
}

func (s tmplStorage) GetTemplates(c context.Context, pageNum, pageSize int64) ([]storage.Template, error) {
	q := s.SelectStruct(templateM{}).Where(sqlx.ColumnDeletedAt.Eq(sqlx.DateTimeZero))
	if pageNum > 0 && pageSize > 0 {
		q.Paginate(pageNum-1, pageSize)
	}

	var _templates []templateM
	err := q.BindRowsContext(c, &_templates)
	if err != nil {
		return nil, err
	}

	templates := make([]storage.Template, len(_templates))
	for i, t := range _templates {
		var args []string
		if t.Args != "" {
			args = strings.Split(t.Args, ",")
		}

		templates[i] = storage.Template{Name: t.Name, Tmpl: t.Tmpl, Args: args}
	}

	return templates, nil
}
