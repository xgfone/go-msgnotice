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
	"fmt"
	"io/ioutil"

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/storage"
	"github.com/xgfone/go-msgnotice/storage/static"
)

// NewChannelStorage returns a new channel storage, which reads all the channels
// from the given file based on the JSON format, for example,
//
//	[
//	    {
//	        "IsDefault": false,
//	        "ChannelName": "xxx",
//	        "DriverName": "yyy",
//	        "DriverConfig": {
//	            // ...
//	        }
//	    }
//	]
func NewChannelStorage(filepath string) (storage.ChannelStorage, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	} else if len(data) == 0 {
		return static.NewChannelStorage(), nil
	}

	var channels []channel.Channel
	if err = json.Unmarshal(data, &channels); err != nil {
		return nil, err
	}

	for _, channel := range channels {
		if channel.DriverName == "" {
			return nil, fmt.Errorf("the driver name of channel named '%s' must not be empty", channel.ChannelName)
		}
	}

	return static.NewChannelStorage(channels...), nil
}
