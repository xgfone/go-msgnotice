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
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/storage"
	"github.com/xgfone/sqlx"
)

// NewChannelStorage returns a new channel storage based on database/sql.
func NewChannelStorage(db *sqlx.DB) storage.ChannelStorage {
	return chStorage{Table: sqlx.NewTable("channel").WithDB(db)}
}

type channelM struct {
	sqlx.Base

	ChannelName string    `sql:"channel_name"`
	DriverName  string    `sql:"driver_name"`
	DriverConf  string    `sql:"driver_conf"`
	IsDefault   sqlx.Bool `sql:"is_default"`
}

type chStorage struct{ sqlx.Table }

func (s chStorage) AddChannel(c context.Context, channel channel.Channel) error {
	var driverConf string
	if len(channel.DriverConf) > 0 {
		buf := bytes.NewBuffer(make([]byte, 0, 128))
		if err := json.NewEncoder(buf).Encode(channel.DriverConf); err != nil {
			return err
		}
		driverConf = buf.String()
	}

	_, err := s.InsertInto().Struct(channelM{
		ChannelName: channel.ChannelName,
		DriverName:  channel.DriverName,
		DriverConf:  driverConf,
		IsDefault:   sqlx.Bool(channel.IsDefault),
	}).ExecContext(c)
	return err
}

func (s chStorage) DelChannel(c context.Context, name string) error {
	_, err := s.Update(sqlx.ColumnDeletedAt.Set(time.Now())).Where(
		sqlx.ColumnDeletedAt.Eq(sqlx.DateTimeZero),
		sqlx.NewColumn("channel_name").Eq(name),
	).ExecContext(c)
	return err
}

func (s chStorage) GetChannel(c context.Context, name string) (channel.Channel, error) {
	var cm channelM
	err := s.SelectStruct(c).Where(
		sqlx.ColumnDeletedAt.Eq(sqlx.DateTimeZero),
		sqlx.NewColumn("channel_name").Eq(name),
	).BindRowStructContext(c, &cm)

	if exist, err := sqlx.CheckErrNoRows(err); err != nil || !exist {
		return channel.Channel{}, err
	}

	channel := channel.Channel{
		ChannelName: cm.ChannelName,
		DriverName:  cm.DriverName,
		IsDefault:   cm.IsDefault.Bool(),
	}
	if cm.DriverConf != "" {
		r := bytes.NewReader([]byte(cm.DriverConf))
		err = json.NewDecoder(r).Decode(&channel.DriverConf)
	}

	return channel, err
}

func (s chStorage) GetChannels(c context.Context, pageNum, pageSize int64) ([]channel.Channel, error) {
	q := s.SelectStruct(channelM{}).Where(sqlx.ColumnDeletedAt.Eq(sqlx.DateTimeZero))
	if pageNum > 0 && pageSize > 0 {
		q.Paginate(pageNum-1, pageSize)
	}

	var _channels []channelM
	err := q.BindRowsContext(c, &_channels)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(_channels, func(i, j int) bool {
		return _channels[i].UpdatedAt.After(_channels[j].UpdatedAt.Time)
	})

	channels := make([]channel.Channel, len(_channels))
	for i, c := range _channels {
		var driverConf map[string]interface{}
		if c.DriverConf != "" {
			err = json.NewDecoder(bytes.NewReader([]byte(c.DriverConf))).Decode(&driverConf)
			if err != nil {
				return nil, err
			}
		}

		channels[i] = channel.Channel{
			ChannelName: c.ChannelName,
			DriverName:  c.DriverName,
			DriverConf:  driverConf,
			IsDefault:   c.IsDefault.Bool(),
		}
	}

	return channels, nil
}
