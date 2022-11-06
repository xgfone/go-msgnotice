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

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/storage"
)

// NewChannelStorage returns a new channel storage.
func NewChannelStorage(channels ...channel.Channel) storage.ChannelStorage {
	s := chStorage{channels: make([]channel.Channel, 0, len(channels))}
	for _, channel := range channels {
		if channel.DriverName == "" {
			panic("the channel driver name must not be empty")
		}
		if s.Index(channel.ChannelName) < 0 {
			s.channels = append(s.channels, channel)
		}
	}
	return s
}

type chStorage struct {
	channels []channel.Channel
}

func (s chStorage) Index(channelName string) (index int) {
	for i, _len := 0, len(s.channels); i < _len; i++ {
		if s.channels[i].ChannelName == channelName {
			return i
		}
	}
	return -1
}

func (s chStorage) AddChannel(context.Context, channel.Channel) error {
	return errors.New("ChannelFileStorage.AddChannel is not implemented")
}

func (s chStorage) DelChannel(c context.Context, name string) error {
	return errors.New("ChannelFileStorage.DelChannel is not implemented")
}

func (s chStorage) GetChannel(c context.Context, name string) (channel.Channel, error) {
	if index := s.Index(name); index > -1 {
		return s.channels[index], nil
	}
	return channel.Channel{}, fmt.Errorf("no channel named '%s'", name)
}

func (s chStorage) GetChannels(c context.Context, pageNum int64, pageSize int64) ([]channel.Channel, error) {
	total := int64(len(s.channels))
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

	channels := make([]channel.Channel, end-start)
	copy(channels, s.channels[start:end])
	return channels, nil
}
