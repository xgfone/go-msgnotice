# A Common Message Notice Library

[![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-msgnotice)](https://pkg.go.dev/github.com/xgfone/go-msgnotice)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-msgnotice/master/LICENSE)
![Minimum Go Version](https://img.shields.io/github/go-mod/go-version/xgfone/go-msgnotice?label=Go%2B)
![Latest SemVer](https://img.shields.io/github/v/tag/xgfone/go-msgnotice?sort=semver)

## Example

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	// Load and register the drivers.
	_ "github.com/xgfone/go-msgnotice/driver/drivers/email"
	_ "github.com/xgfone/go-msgnotice/driver/drivers/feishu"

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/channel/manager"
	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/builder"
	"github.com/xgfone/go-msgnotice/driver/middleware"
	"github.com/xgfone/go-msgnotice/driver/middleware/logger"
	"github.com/xgfone/go-msgnotice/driver/middleware/timeout"
)

var (
	listenaddr   = flag.String("listenaddr", "127.0.0.1:80", "The address to listen on.")
	channelsfile = flag.String("channelsfile", "channels.json", "The file path storing the channels.")
)

func _loadFromFile(filepath string, dst any, cb func() error) (err error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	} else if len(data) == 0 {
		return
	}

	if err = json.Unmarshal(data, dst); err != nil {
		return
	}

	return cb()
}

func initChannels() {
	type Channel struct {
		ChannelName string
		DriverName  string
		DriverConf  map[string]any
	}

	var channels []Channel
	var _channels []channel.Channel
	err := _loadFromFile(*channelsfile, &channels, func() (err error) {
		for _, c := range channels {
			if c.ChannelName == "" || c.DriverName == "" {
				return errors.New("missing the channel or driver name")
			}

			driver, err := builder.Build(c.DriverName, c.DriverConf)
			if err != nil {
				return err
			}

			_channels = append(_channels, channel.New(c.ChannelName, driver))
		}
		return
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	manager.Default.UpsertChannels(_channels...)
}

func initMiddlewares() {
	middleware.DefaultManager.Use(timeout.New(100, time.Second*10, nil))
	middleware.DefaultManager.Use(logger.New(90, nil))
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content  any
		Channel  string
		Receiver string
		Metadata map[string]any
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Content == nil || req.Channel == "" || req.Receiver == "" {
		http.Error(w, "missing the required arguments", 400)
		return
	}

	msg := driver.NewMessage(req.Channel, "", req.Receiver, req.Content, req.Metadata)
	err := channel.Send(context.Background(), msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	flag.Parse()

	initMiddlewares()
	initChannels()

	http.HandleFunc("/message", httpHandler)
	_ = http.ListenAndServe(*listenaddr, nil)
}
```

```json
// Channels configuration file: channels.json
[
  {
    "ChannelName": "email",
    "DriverName": "email",
    "DriverConf": {
      "addr": "smtp.example.com",
      "from": "username@example.com",
      "username": "username@example.com",
      "password": "password",
      "forcetls": true
    }
  },
  {
    "ChannelName": "feishu.webhook",
    "DriverName": "feishu.webhook"
  }
]
```

```bash
# Run the message notice process.
$ msgnotice &

# Client sends the message.

## Send the message by the email
$ curl http://localhost/message -X POST -H 'content-type: application/json' \
-d '{"Content": {"Subject": "subject", "Content": "test"}, "Channel": "email", "Receiver": "xiegaofeng@bimoai.com"}'

## Send the general text message to feishu by webhook.
$ curl http://localhost/message -X POST -H 'content-type: application/json' \
-d '{"Channel": "feishu.webhook", "Receiver": "9c473cb6-1234-5678-bbf1-147b5e83ab8e", "Content": "<at user_id=\"all\">All</at>"}'

## Send the rich text message to feishu by webhook.
$ curl http://localhost/message -X POST -H 'content-type: application/json' -d '{
    "Channel": "feishu.webhook",
    "Receiver": "9c473cb6-1234-5678-bbf1-147b5e83ab8e",
    "Metadata": {"MsgType": "post"},
    "Content": {
        "zh_cn": {
            "title": "test",
            "content": [
                [
                    {"tag": "text", "text": "first paragraph"},
                    {"tag": "at", "user_id": "all"},
                    {"tag": "a", "text": "URL", "href": "https://www.example.com"}
                ],
                [
                    {"tag": "text", "text": "second paragraph"}
                ]
            ]
        }
    }
}'
```
