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
)

var (
	listenaddr    = flag.String("listenaddr", "127.0.0.1:80", "The address to listen on.")
	channelsfile  = flag.String("channelsfile", "channels.json", "The file path storing the channels.")
	templatesfile = flag.String("templatesfile", "templates.json", "The file path storing the templates.")
)

func wrapDriver(_name, _type string, _driver driver.Driver) driver.Driver {
	return driver.MatchAndWrap(_type, _driver, func(c context.Context, m driver.Message, d driver.Driver) error {
		log.Printf("middleware=%s, type=%s, title=%s, content=%s, receiver=%s",
			_name, _type, m.Title, m.Content, m.Receiver)
		return _driver.Send(c, m)
	})
}

func newDriverMiddleware(_name, _type string, _prio int) middleware.Middleware {
	return middleware.New(_name, _prio, func(d driver.Driver) driver.Driver {
		return wrapDriver(_name, _type, d)
	})
}

func _loadFromFile(filepath string, dst interface{}, cb func() error) (err error) {
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

func initDriverMiddlewares() (err error) {
	var templates []template.Template
	err = _loadFromFile(*templatesfile, &templates, func() error {
		for _, tmpl := range templates {
			if tmpl.Name == "" || tmpl.Tmpl == "" {
				return errors.New("template misses the name or content")
			}
		}
		return nil
	})
	if err != nil {
		return
	}

	getTmpl := func(c context.Context, name string) (t template.Template, ok bool, err error) {
		for _, tmpl := range templates {
			if tmpl.Name == name {
				return tmpl, true, nil
			}
		}
		err = fmt.Errorf("no template named '%s'", name)
		return
	}

	// For Common middlewares
	middleware.DefaultManager.Use(logger.New(0, ""), template.New(10, "", template.GetterFunc(getTmpl)))

	// Only for Email middleware
	middleware.DefaultManager.Use(newDriverMiddleware("email", "email", 20))

	// Only for Stdout middleware
	middleware.DefaultManager.Use(newDriverMiddleware("stdout", "stdout", 20))

	return nil
}

func initChannels() (err error) {
	var channels []channel.Channel
	err = _loadFromFile(*channelsfile, &channels, func() (err error) {
		for i, channel := range channels {
			channels[i], err = channel.Init()
			if err != nil {
				return
			}
		}
		return
	})

	if err == nil {
		manager.Default.UpsertChannels(channels...)
	}
	return
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

	msg := driver.NewMessage("", req.Receiver, req.Title, req.Content, req.Metadata)
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
	_ = http.ListenAndServe(*listenaddr, nil)
}
```

```json
// Templates configuration file: templates.json
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
