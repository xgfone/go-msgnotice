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
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/knadh/smtppool"
	"github.com/xgfone/go-msgnotice/driver"
	"github.com/xgfone/go-msgnotice/driver/builder"
)

// DriverType represents the driver type "email".
const DriverType = "email"

func init() { builder.NewAndRegister(DriverType, New) }

// New returns a new driver, which sends the message by the html email,
// which is registered as the driver builder with name "email"
// and type DriverType by default.
//
// config options:
//
//	addr(string, required): the mail server address, such as "mail.examole.com".
//	from(string, required): the adddress to send email, such as "username@mail.example.com".
//	username(string, required): the username to login the mail server, such as "username@mail.example.com".
//	password(string, required): the password to login the mail server, such as "password".
//	forcetls(int|int64|uint|uint64|string|bool, optional): if true, force to use TLS. For integer, 0 is false else true.
//	timeout(int|int64|uint|uint64|string, optional): the timeout. If integer, stand for second. default 3s.
//	idletimeout(int|int64|uint|uint64|string, optional): time idle timeout. If integer, stand for second. default 1m.
//	maxconnnum(int|int64|uint|uint64, optional): the maximum number of the connection, default 100.
//
// If addr does not contain the port, use 465 if forcetls is true else 25.
//
// Notice: The returned driver supports the comma-separated receiver list.
func New(config map[string]any) (driver.Driver, error) {
	addr, _ := config["addr"].(string)
	from, _ := config["from"].(string)
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)
	if addr == "" {
		return nil, errors.New("addr is missing or invalid")
	}
	if from == "" {
		return nil, errors.New("from is missing or invalid")
	}
	if username == "" {
		return nil, errors.New("username is missing or invalid")
	}
	if password == "" {
		return nil, errors.New("password is missing or invalid")
	}

	var maxconnnum int
	switch v := config["maxconnnum"].(type) {
	case nil:
		maxconnnum = 100
	case int:
		maxconnnum = v
	case int64:
		maxconnnum = int(v)
	case uint:
		maxconnnum = int(v)
	case uint64:
		maxconnnum = int(v)
	default:
		return nil, fmt.Errorf("unsupported maxconnnum type %T", v)
	}

	var timeout time.Duration
	switch v := config["timeout"].(type) {
	case nil:
		timeout = 3 * time.Second

	case int:
		timeout = time.Duration(v) * time.Second
	case int64:
		timeout = time.Duration(v) * time.Second
	case uint:
		timeout = time.Duration(v) * time.Second
	case uint64:
		timeout = time.Duration(v) * time.Second

	case string:
		t, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %s", err)
		}
		timeout = t

	default:
		return nil, fmt.Errorf("unsupported timeout type %T", v)
	}

	var idleTimeout time.Duration
	switch v := config["idletimeout"].(type) {
	case nil:
		idleTimeout = time.Minute

	case int:
		idleTimeout = time.Duration(v) * time.Second
	case int64:
		idleTimeout = time.Duration(v) * time.Second
	case uint:
		idleTimeout = time.Duration(v) * time.Second
	case uint64:
		idleTimeout = time.Duration(v) * time.Second

	case string:
		t, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid idle timeout: %s", err)
		}
		idleTimeout = t

	default:
		return nil, fmt.Errorf("unsupported idle timeout type %T", v)
	}

	var forceTLS bool
	switch v := config["forcetls"].(type) {
	case nil:
	case bool:
		forceTLS = v

	case int:
		forceTLS = v != 0
	case int64:
		forceTLS = v != 0
	case uint:
		forceTLS = v != 0
	case uint64:
		forceTLS = v != 0

	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("invalid forcetls: %s", err)
		}
		forceTLS = b

	default:
		return nil, fmt.Errorf("unsupported forcetls type %T", v)
	}

	var port int
	var host string
	if _host, _port, err := net.SplitHostPort(addr); err == nil {
		host = _host
		v, err := strconv.ParseUint(_port, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
		port = int(v)
	} else {
		host = addr
		if forceTLS {
			port = 465
		} else {
			port = 25
		}
	}

	var auth smtp.Auth
	if forceTLS {
		auth = smtp.PlainAuth("", username, password, host)
	} else {
		auth = newPlainAuth("", username, password, host)
	}

	opt := smtppool.Opt{
		Host:            host,
		Port:            port,
		Auth:            auth,
		MaxConns:        maxconnnum,
		IdleTimeout:     idleTimeout,
		PoolWaitTimeout: timeout,

		TLSConfig: nil,
		SSL:       forceTLS,
	}
	pool, err := smtppool.New(opt)
	if err != nil {
		return nil, err
	}

	return driverImpl{from: from, pool: pool}, nil
}

type driverImpl struct {
	pool *smtppool.Pool
	from string
}

func (d driverImpl) Type() string { return DriverType }

func (d driverImpl) Stop() { d.pool.Close() }

func (d driverImpl) Send(c context.Context, m driver.Message) error {
	var mail smtppool.Email
	mail.From = d.from
	mail.To = strings.Split(m.Receiver, ",")
	mail.Subject = m.Title
	mail.HTML = []byte(m.Content)
	return d.pool.Send(mail)
}

type plainAuth struct {
	identity, username, password string
	host                         string
}

func newPlainAuth(identity, username, password, host string) smtp.Auth {
	return &plainAuth{identity, username, password, host}
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
