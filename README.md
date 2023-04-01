# A Common Message Notice Library [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-msgnotice)](https://pkg.go.dev/github.com/xgfone/go-msgnotice) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-msgnotice/master/LICENSE)

A common message notice library supporting `Go1.18+`.


## Example
```go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Load and register the drivers.
	_ "github.com/xgfone/go-msgnotice/driver/drivers/email"
	_ "github.com/xgfone/go-msgnotice/driver/drivers/stdout" // For test

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/channel/manager"
	"github.com/xgfone/go-msgnotice/driver"
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

func wrapDriver(_name, _type string, _driver driver.Driver) driver.Driver {
	return driver.NewDriver(func(c context.Context, m driver.Message) error {
		log.Printf("middleware=%s, type=%s, title=%s, content=%s, receiver=%s",
			_name, _type, m.Title, m.Content, m.Receiver)
		return _driver.Send(c, m)
	}, _driver.Stop)
}

func newDriverMiddleware(_name, _type string, _prio int) middleware.Middleware {
	return middleware.NewMiddleware(_name, _type, _prio, func(d driver.Driver) driver.Driver {
		return wrapDriver(_name, _type, d)
	})
}

func initDriverMiddlewares() (err error) {
	tmplStorage, err := file.NewTemplateStorage(*templatesfile)
	if err != nil {
		return
	}

	// For Common middlewares
	middleware.DefaultManager.Use(logger.New("", 0), template.New("", 10, tmplStorage.GetTemplate))

	// Only for Email middleware
	middleware.DefaultManager.Use(newDriverMiddleware("email", "email", 20))

	// Only for Stdout middleware
	middleware.DefaultManager.Use(newDriverMiddleware("stdout", "stdout", 20))

	return nil
}

func initChannels() (err error) {
	channelStorage, err := file.NewChannelStorage(*channelsfile)
	if err != nil {
		return
	}

	channels, _ := channelStorage.GetChannels(context.Background(), 0, 0)
	return manager.Default.BuildAndUpsertChannels(channels...)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Timeout  string
		Channel  string
		Title    string
		Content  string
		Receiver string
		Metadata map[string]interface{}
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Channel == "" || req.Title == "" || req.Content == "" || len(req.Receiver) == 0 {
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

	msg := driver.NewMessage(req.Receiver, req.Title, req.Content, req.Metadata)
	err := channel.Send(ctx, req.Channel, msg)
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
            "addr": "mail.domain.com:25",
            "from": "username@domain.com",
            "username": "username@domain.com",
            "password": "password"
        }
    }
]
```

```bash
# Run the message notice process.
$ msgnotice &

# Client sends the message.
$ curl http://localhost/message -XPOST -H 'Content-Type: application/json' \
-d '{"Channel":"stdout", "Title":"title", "Content":"content", "Receiver":"someone"}'
$ curl http://localhost/message -XPOST -H 'Content-Type: application/json' \
-d '{"Channel":"email", "Title":"title", "Content":"content", "Receiver":"someone@mail.com"}'
$ curl http://localhost/message -XPOST -H 'Content-Type: application/json' \
-d '{"Channel":"email", "Title":"title", "Content":"tmpl:hello", "Metadata":{"name":"xgfone"}, "Receiver":"someone@mail.com"}'
```
