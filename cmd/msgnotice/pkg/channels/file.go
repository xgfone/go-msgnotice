// Copyright 2023 xgfone
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

// Package channels is used to load the channels.
package channels

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/channel/manager"
)

// The channel file format:
//
//	[
//	    {
//	        "ChannelName": "xxx",
//	        "DriverName": "yyy",
//	        "DriverConfig": {
//	            // ...
//	        }
//	    }
//	]
var channelsfile = gconf.NewString("channelsfile", "channels.json",
	"The json file path storing the channel configs.")

func init() {
	atexit.OnInit(func() {
		chfile := channelsfile.Get()
		channels, err := _loadChannelsFromFile(chfile)
		if err != nil {
			log.Fatal("fail to load channels from file", "file", chfile, "err", err)
		}

		err = manager.Default.BuildAndUpsertChannels(channels...)
		if err != nil {
			log.Fatal("fail to build the channels", "channels", channels, "err", err)
		}
	})
}

func _loadChannelsFromFile(filepath string) (channels []channel.Channel, err error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	} else if len(data) == 0 {
		return
	}

	if err = json.Unmarshal(data, &channels); err != nil {
		return
	}

	for _, channel := range channels {
		if channel.DriverName == "" {
			err = fmt.Errorf("the driver name of channel named '%s' must not be empty", channel.ChannelName)
			return
		}
	}

	return
}
