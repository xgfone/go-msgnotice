// Copyright 2025 xgfone
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

package feishu

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/xgfone/go-toolkit/jsonx"
	"github.com/xgfone/go-toolkit/unsafex"
)

const webhookBaseURL = "https://open.feishu.cn/open-apis/bot/v2/hook/"

// Webhook is a webhook to send the feishu message.
type Webhook struct {
	do func(*http.Request) (*http.Response, error)

	id  string
	key string
	url string
}

// NewWebhook returns a new Webhook.
//
// botid is required. but secret is optional.
// If secret is not empty, enable the signature verification.
func NewWebhook(botid, secret string) Webhook {
	return Webhook{}.WithBot(botid, secret).WithSender(http.DefaultClient.Do)
}

// WithSender returns a new Webhook with the http sender.
//
// Default: http.DefaultClient.Do
func (w Webhook) WithSender(do func(*http.Request) (*http.Response, error)) Webhook {
	if do == nil {
		panic("Webhook.WithSender: do is nil")
	}

	w.do = do
	return w
}

// WithURL returns a new Webhook with the bot id and secret.
//
// botid is required. but secret is optional.
// If secret is not empty, enable the signature verification.
func (w Webhook) WithBot(botid, secret string) Webhook {
	if botid == "" {
		panic("Webhook.WithBot: botid is empty")
	}

	w.id = botid
	w.key = secret
	w.url = webhookBaseURL + botid
	return w
}

// SendText sends a plain text message.
//
// @all: <at user_id="all">All</at>
//
// @single_user: <at user_id="open_id or user_id">Name</at>
//
// See https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot#756b882f
func (w Webhook) SendText(ctx context.Context, text string) (err error) {
	return w.send(ctx, "text", map[string]string{"text": text})
}

// SendRich sends a rich text message.
//
// See https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot#f62e72d5
func (w Webhook) SendRich(ctx context.Context, content any) (err error) {
	return w.send(ctx, "post", map[string]any{"post": content})
}

// SendRichZh is short for SendRich to send a rich text message with the title and chinese content.
//
// content is an array, each element of whose is a paragraph. The format is like:
//
//	[]any{
//		// First Paragraph
//		[]any{
//			map[string]any{"tag": "text", "text": "文本"},
//			map[string]any{"tag": "a", "text": "描述信息", "href": "http://..."},
//			map[string]any{"tag": "at", "user_id": "open_id or user_id"},
//		},
//
//		// Second Paragraph
//		[]any{
//			map[string]any{ /*...*/ },
//			map[string]any{ /*...*/ },
//		},
//
//		// ...
//	}
func (w Webhook) SendRichZh(ctx context.Context, title string, content any) (err error) {
	content = map[string]any{"title": title, "content": content}
	return w.send(ctx, "post", map[string]any{"post": map[string]any{"zh_cn": content}})
}

// SendRichEn is short for SendRich to send a rich text message with the title and english content.
//
// content is an array, each element of whose is a paragraph. The format is like:
//
//	[]any{
//		// First Paragraph
//		[]any{
//			map[string]any{"tag": "text", "text": "文本"},
//			map[string]any{"tag": "a", "text": "描述信息", "href": "http://..."},
//			map[string]any{"tag": "at", "user_id": "open_id or user_id"},
//		},
//
//		// Second Paragraph
//		[]any{
//			map[string]any{ /*...*/ },
//			map[string]any{ /*...*/ },
//		},
//
//		// ...
//	}
func (w Webhook) SendRichEn(ctx context.Context, title string, content any) (err error) {
	content = map[string]any{"title": title, "content": content}
	return w.send(ctx, "post", map[string]any{"post": map[string]any{"en_us": content}})
}

func (w Webhook) send(ctx context.Context, msgtype string, content any) (err error) {
	type (
		Request struct {
			Timestamp string `json:"timestamp"`
			Sign      string `json:"sign"`

			MsgType string `json:"msg_type"`
			Content any    `json:"content"`
		}

		Response struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data any    `json:"data"`
		}
	)

	req := Request{MsgType: msgtype, Content: content}
	req.Sign, req.Timestamp = w.getsign()

	buf := getbuffer()
	defer putbuffer(buf)
	if err = jsonx.Marshal(buf, req); err != nil {
		return fmt.Errorf("fail to encode message by json: %w", err)
	}

	httpreq, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, buf)
	if err != nil {
		return fmt.Errorf("fail to create a new request: %w", err)
	}
	httpreq.Header.Set("Content-Type", "application/json")

	httpresp, err := w.do(httpreq)
	if err != nil {
		return fmt.Errorf("fail to send the request: %w", err)
	}
	defer httpresp.Body.Close()

	data, err := io.ReadAll(httpresp.Body)
	if err != nil {
		return fmt.Errorf("fail to read the response body: %w", err)
	}

	var resp Response
	if err = jsonx.Unmarshal(&resp, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("fail to decode the response body by json: data=%s, err=%w", unsafex.String(data), err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("%d: %s", resp.Code, resp.Msg)
	}

	return
}

func (w Webhook) getsign() (sign, timestamp string) {
	if w.key == "" {
		return
	}

	timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	signedstr := fmt.Sprintf("%v\n%s", timestamp, w.key)
	hmac := hmac.New(sha256.New, unsafex.Bytes(signedstr))
	sign = base64.StdEncoding.EncodeToString(hmac.Sum(nil))
	return
}

var bufpool = sync.Pool{New: func() any { return bytes.NewBuffer(make([]byte, 0, 1024)) }}

func getbuffer() *bytes.Buffer  { return bufpool.Get().(*bytes.Buffer) }
func putbuffer(b *bytes.Buffer) { b.Reset(); bufpool.Put(b) }
