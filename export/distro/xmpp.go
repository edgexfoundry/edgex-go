//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"crypto/tls"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/mattn/go-xmpp"
	"strings"
	"time"
)

type xmppSender struct {
	client  *xmpp.Client
	remote  string
	msgType string
	subject string
	thread  string
	other   []string
	stamp   time.Time
}

func NewXMPPSender(addr models.Addressable) Sender {
	protocol := strings.ToLower(addr.Protocol)

	if protocol == "tls" {
		xmpp.DefaultConfig = tls.Config{
			ServerName:         serverName(addr.Address),
			InsecureSkipVerify: false,
		}
	}

	options := xmpp.Options{
		Host:     addr.Address,
		User:     addr.User,
		Password: addr.Password,
		NoTLS:    protocol == "tls",
		Debug:    false,
		Session:  false,
	}

	xmppClient, err := options.NewClient()
	if err != nil {
		logger.Error(err.Error())
	}

	sender := &xmppSender{
		client: xmppClient,
	}

	return sender
}

func (sender *xmppSender) Send(data []byte) {
	stringData := string(data)

	sender.client.Send(xmpp.Chat{
		Text:    stringData,
		Remote:  sender.remote,
		Subject: sender.subject,
		Thread:  sender.thread,
		Other:   sender.other,
		Stamp:   sender.stamp,
	})
}

func serverName(host string) string {
	return strings.Split(host, ":")[0]
}
