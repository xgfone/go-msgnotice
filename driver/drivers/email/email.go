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

// Package email provides a driver to send the message by the email.
package email

import (
	"context"
	"errors"
	"net"
	"net/smtp"

	"github.com/jordan-wright/email"
	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/builder"
)

func init() { builder.NewAndRegister("email", "email", New) }

// New returns a new driver, which sends the message by the email.
func New(config map[string]interface{}) (driver.Driver, error) {
	addr, _ := config["Addr"].(string)
	from, _ := config["From"].(string)
	username, _ := config["Username"].(string)
	password, _ := config["Password"].(string)
	forceTLS, _ := config["ForceTLS"].(bool)
	if addr == "" {
		return nil, errors.New("Addr is missing or invalid")
	}
	if from == "" {
		return nil, errors.New("From is missing or invalid")
	}
	if username == "" {
		return nil, errors.New("Username is missing or invalid")
	}
	if password == "" {
		return nil, errors.New("Password is missing or invalid")
	}

	hostname := addr
	if host, _, err := net.SplitHostPort(addr); err == nil {
		hostname = host
	}

	var auth smtp.Auth
	if forceTLS {
		auth = smtp.PlainAuth("", username, password, hostname)
	} else {
		auth = newPlainAuth("", username, password, hostname)
	}

	pool, err := email.NewPool(addr, 100, auth)
	if err != nil {
		return nil, err
	}

	return driverImpl{from: from, pool: pool}, nil
}

type driverImpl struct {
	pool *email.Pool
	from string
}

func (d driverImpl) Stop() { d.pool.Close() }
func (d driverImpl) Send(c context.Context, title, content string,
	metadata map[string]interface{}, tos ...string) error {
	mail := email.NewEmail()
	mail.From = d.from
	mail.To = tos
	mail.Subject = title
	mail.HTML = []byte(content)
	return d.pool.Send(mail, -1)
}

type plainAuth struct {
	identity, username, password string
	host                         string
}

func newPlainAuth(identity, username, password, host string) smtp.Auth {
	return &plainAuth{identity, username, password, host}
}

func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}

func (a *plainAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *plainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}
