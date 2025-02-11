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

// Package feishu provides a driver to send the message to feishu.
package feishu

import (
	"context"
	"fmt"

	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/builder"
	"github.com/xgfone/go-msgnotice/tools/feishu"
)

// DriverTypeWebhook represents the driver type "feishu.webhook".
const DriverTypeWebhook = "feishu.webhook"

func init() { builder.NewAndRegister(DriverTypeWebhook, NewWebhook) }

// NewWebhook returns a new driver, which sends the message to feishu.
//
// The optional secret is extracted from the message metadata by name "secret".
func NewWebhook(config map[string]any) (driver.Driver, error) {
	_secrets := config["secrets"].(map[string]any)
	secrets := make(map[string]string, len(_secrets))
	for k, v := range _secrets {
		var ok bool
		if secrets[k], ok = v.(string); !ok {
			return nil, fmt.Errorf("secret of '%s' is %T, not string", k, v)
		}
	}

	return driver.New(DriverTypeWebhook, func(c context.Context, m driver.Message) error {
		secret, ok := m.Metadata["secret"].(string)
		if !ok {
			secret = secrets[m.Receiver]
		}
		return feishu.NewWebhook(m.Receiver, secret).SendText(c, m.Content)
	}, nil), nil
}
