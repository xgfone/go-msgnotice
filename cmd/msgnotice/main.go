package main

import (
	"context"
	"time"

	_ "github.com/xgfone/go-msgnotice/cmd/msgnotice/pkg/appinit"
	_ "github.com/xgfone/go-msgnotice/cmd/msgnotice/pkg/channels"
	_ "github.com/xgfone/go-msgnotice/cmd/msgnotice/pkg/drivers"
	_ "github.com/xgfone/go-msgnotice/cmd/msgnotice/pkg/middlewares"
	_ "github.com/xgfone/goapp/log/middleware"

	"github.com/xgfone/go-apiserver/entrypoint"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/http/router/ruler"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-atexit"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/goapp"

	"github.com/xgfone/go-msgnotice/channel"
	"github.com/xgfone/go-msgnotice/channel/manager"
	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/middleware/logger"
)

var listenaddr = gconf.NewString("listenaddr", ":80", "The address to listen to.")

func main() {
	goapp.Init("")
	entrypoint.Start(listenaddr.Get(), nil)
}

func init() {
	logger.LogEvent = logEvent
	atexit.OnExit(manager.Default.Stop)
	ruler.DefaultRouter.Path("/msgnotice").POSTContextWithError(sendMsgNoticeHandler)
}

func sendMsgNoticeHandler(c *reqresp.Context) (err error) {
	var req struct {
		Title    string `validate:"required"`
		Content  string `validate:"required"`
		Channel  string `validate:"required"`
		Receiver string `validate:"required"`
		Metadata map[string]interface{}
		Timeout  int64 // Unit: ms
	}
	if err = c.BindBody(&req); err != nil {
		return c.Text(400, err.Error())
	}

	ctx := c.Context()
	if req.Timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, time.Millisecond*time.Duration(req.Timeout))
		defer cancel()
	}

	msg := driver.NewMessage(req.Receiver, req.Title, req.Content, req.Metadata)
	return channel.Send(ctx, req.Channel, msg)
}

func logEvent(e logger.Event) {
	var cname string
	if e.Channel != nil {
		cname = e.Channel.ChannelName
	}

	if e.Err == nil {
		log.Info("send message notice", "channel", cname, "title", e.Title,
			"content", e.Content, "metadata", e.Metadata, "receiver", e.Receiver,
			"start", e.Start.Unix(), "cost", time.Since(e.Start))
	} else {
		log.Error("send message notice", "channel", cname, "title", e.Title,
			"content", e.Content, "metadata", e.Metadata, "receiver", e.Receiver,
			"start", e.Start.Unix(), "cost", time.Since(e.Start), "err", e.Err)
	}
}
