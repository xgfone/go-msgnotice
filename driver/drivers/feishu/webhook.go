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
	"context"
	"fmt"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/builder"
	"github.com/xgfone/go-msgnotice/tools/feishu"
	"github.com/xgfone/go-toolkit/mapx"
)

// DriverTypeWebhook represents the driver type "feishu.webhook".
const DriverTypeWebhook = "feishu.webhook"

func init() { builder.NewAndRegister(DriverTypeWebhook, buildWebhook) }

func buildWebhook(name string, config map[string]any) (driver.Driver, error) {
	var secrets map[string]string
	switch _secrets := config["Secrets"].(type) {
	case nil:
	case map[string]any:
		secrets = mapx.Convert(_secrets, func(k string, v any) (string, string) { return k, v.(string) })

	default:
		return nil, fmt.Errorf("expect Secrets is a map[string]any, but got %T", _secrets)
	}

	return NewWebhook(name).WithLookup(func(receiver string) (string, error) {
		return secrets[receiver], nil
	}), nil
}

/// ---------------------------------------------------------------------- ///

func noop(string) (string, error) { return "", nil }

var _ driver.Driver = Webhook{}

// Webhook is a driver to send the message to feishu by webhook.
type Webhook struct {
	lookup func(receiver string) (secret string, err error)
	name   string
}

// NewWebhook returns a new driver based on feishu webhook.
func NewWebhook(name string) Webhook {
	return Webhook{name: name}.WithLookup(noop)
}

// WithLookup returns a new feishu webhook driver with the secret lookup function.
func (w Webhook) WithLookup(f func(receiver string) (secret string, err error)) Webhook {
	if f == nil {
		panic("driver.feishu.Webhook: the secret lookup function is nil")
	}

	w.lookup = f
	return w
}

// Send implements the interface driver.Driver#Stop.
func (w Webhook) Stop() {}

// Name implements the interface driver.Driver#Name.
func (w Webhook) Name() string { return w.name }

// Type implements the interface driver.Driver#Type.
func (w Webhook) Type() string { return DriverTypeWebhook }

// Send implements the interface driver.Driver#Send.
func (w Webhook) Send(c context.Context, m driver.Message) (err error) {
	secret, err := w.lookup(m.Receiver)
	if err != nil {
		return err
	}

	webhook := feishu.NewWebhook(m.Receiver, secret)
	msgtype, _ := m.Metadata["MsgType"].(string)
	switch msgtype {
	case "", "text":
		if content, ok := m.Content.(string); ok {
			err = webhook.SendText(c, content)
		} else {
			err = fmt.Errorf("expect the content is a string, but got %T", m.Content)
		}

	case "post":
		err = webhook.SendRich(c, m.Content)

	default:
		err = fmt.Errorf("driver.feishu.webhook: unkown msg type '%s'", msgtype)
	}

	return
}
