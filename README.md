# A Common Message Notice Library [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-msgnotice)](https://pkg.go.dev/github.com/xgfone/go-msgnotice) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-msgnotice/master/LICENSE)

A common message notice library supporting `Go1.11+`.


## Example
```go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	// Load and register the drivers.
	_ "github.com/xgfone/go-msgnotice/driver/drivers/email"
	_ "github.com/xgfone/go-msgnotice/driver/drivers/stdout" // For test

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/channel/manager"
	"github.com/xgfone/go-msgnotice/driver/middleware"
	"github.com/xgfone/go-msgnotice/driver/middleware/logger"
	"github.com/xgfone/go-msgnotice/driver/middleware/template"
	"github.com/xgfone/go-msgnotice/storage/file"
)

var (
	listenaddr    = flag.String("listenaddr", ":80", "The address to listen on.")
	channelsfile  = flag.String("channelsfile", "channels.json", "The file path storing the channels.")
	templatesfile = flag.String("templatesfile", "templates.json", "The file path storing the templates.")
)

func newChannel(cname, dname string, dconf map[string]interface{}) (*channel.Channel, error) {
	channel, err := channel.NewChannel(cname, dname, dconf)
	if err == nil {
		channel.Driver = middleware.DefaultManager.WrapDriver(channel.Driver)
	}
	return channel, err
}

func initChannels() (err error) {
	manager.Default.NewChannel = newChannel
	channelStorage, err := file.NewChannelStorage(*channelsfile)
	if err != nil {
		return
	}

	channels, _ := channelStorage.GetChannels(context.Background(), 0, 0)
	return manager.Default.BuildAndUpsertChannels(channels...)
}

func initDriverMiddlewares() (err error) {
	tmplStorage, err := file.NewTemplateStorage(*templatesfile)
	if err != nil {
		return
	}

	middleware.DefaultManager.Use(logger.New(0), template.New(10, tmplStorage.GetTemplate))
	return nil
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Timeout   string
		Channel   string
		Title     string
		Content   string
		Metadata  map[string]interface{}
		Receivers []string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Channel == "" || req.Title == "" || req.Content == "" || len(req.Receivers) == 0 {
		http.Error(w, "missing the required arguments", 400)
		return
	}

	ctx := context.Background()
	if req.Timeout != "" {
		timeout, err := time.ParseDuration(req.Timeout)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	err := channel.Send(ctx, req.Channel, req.Title, req.Content, req.Metadata, req.Receivers...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	flag.Parse()

	if err := initDriverMiddlewares(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := initChannels(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.HandleFunc("/message", httpHandler)
	http.ListenAndServe(*listenaddr, nil)
}
```

```json
// Templates configuration file: template.json
[
    {
        "Name": "hello",
        "Tmpl": "Hello {name}",
        "Args": ["name"]
    }
]
```

```json
// Channels configuration file: channels.json
[
    {
        "ChannelName": "stdout",
        "DriverName": "stdout"
    },
    {
        "ChannelName": "email",
        "DriverName": "email",
        "DriverConf": {
            "Addr": "mail.domain.com:25",
            "From": "username@domain.com",
            "Username": "username@domain.com",
            "Password": "password"
        }
    }
]
```

```bash
# Run the message notice process.
$ msgnotice &

# Client sends the message.
$ curl http://localhost/message -XPOST -H 'Content-Type: application/json'\
-d '{"Channel":"stdout", "Title":"title", "Content":"content", "Receivers":["someone"]}'
$ curl http://localhost/message -XPOST -H 'Content-Type: application/json'\
-d '{"Channel":"email", "Title":"title", "Content":"content", "Receivers":["someone@mail.com"]}'
$ curl http://localhost/message -XPOST -H 'Content-Type: application/json'\
-d '{"Channel":"email", "Title":"title", "Content":"tmpl:hello", "Metadata":{"name":"xgfone"}, "Receivers":["someone@mail.com"]}'
```
