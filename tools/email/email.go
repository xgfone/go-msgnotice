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

// Package email privides some functions to send the feishu messages.
package email

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xgfone/go-toolkit/jsonx"
)

// Message represents the email message.
type Message struct {
	Subject string
	Content string
}

// DecodeMessage deocdes the message content and returns the subject and content.
func DecodeMessage(msgContent any) (subject, content string, err error) {
	type messager interface {
		Message() (subject, content string)
	}

	switch v := msgContent.(type) {
	case Message:
		subject, content = v.Subject, v.Content

	case messager:
		subject, content = v.Message()

	case interface{ Message() Message }:
		m := v.Message()
		subject, content = m.Subject, m.Content

	case map[string]any:
		var ok bool
		if subject, ok = v["Subject"].(string); !ok {
			err = fmt.Errorf("driver.email: 'Subject' expects a string, but got %T", v["Subject"])
			return
		}
		if content, ok = v["Content"].(string); !ok {
			err = fmt.Errorf("driver.email: 'Content' expects a string, but got %T", v["Content"])
			return
		}

	case []byte:
		var m Message
		err = jsonx.Unmarshal(&m, bytes.NewReader(v))
		subject, content = m.Subject, m.Content

	case json.RawMessage:
		var m Message
		err = jsonx.Unmarshal(&m, bytes.NewReader(v))
		subject, content = m.Subject, m.Content

	default:
		err = fmt.Errorf("driver.email: unsupported content type %T", v)
	}

	return
}
