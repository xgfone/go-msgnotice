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

import (
	"context"

	"github.com/xgfone/go-msgnotice/channel"
)

// ChannelStorage is used to manage the channel in the stroage.
type ChannelStorage interface {
	// If the channel has exists by the name, return an error.
	AddChannel(context.Context, channel.Channel) error

	// Return nil if the channel named name does not exist.
	DelChannel(c context.Context, name string) error

	// Return ZERO if the channel named name does not exist.
	GetChannel(c context.Context, name string) (channel.Channel, error)

	// If any of pageNum and pageSize is less than or equal to 0,
	// return all the channels. pageNum starts with 1.
	GetChannels(c context.Context, pageNum, pageSize int64) ([]channel.Channel, error)
}
