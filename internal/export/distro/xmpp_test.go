//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"crypto/tls"
	"flag"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/mattn/go-xmpp"
)

const (
	TestMessage = "hello world"
)

var server = flag.String("server", "talk.google.com:443", "server")

//your gmail account, eg: xxx@gmail.com
var username = flag.String("username", "", "username")

//your Gmail password
var password = flag.String("password", "", "password")
var status = flag.String("status", "xa", "status")
var statusMessage = flag.String("status-msg", "I for one welcome our new codebot overlords.", "status message")
var notls = flag.Bool("notls", false, "No TLS")
var debug = flag.Bool("debug", false, "debug output")
var session = flag.Bool("session", false, "use server session")

//if you want to test, replace this value with `true`
var testFlag = false

func getServerName(host string) string {
	return strings.Split(host, ":")[0]
}

func TestXmppSend(t *testing.T) {
	if testFlag {
		if !*notls {
			xmpp.DefaultConfig = tls.Config{
				ServerName:         getServerName(*server),
				InsecureSkipVerify: false,
			}
		}

		var talk *xmpp.Client
		var err error
		options := xmpp.Options{Host: *server,
			User:          *username,
			Password:      *password,
			NoTLS:         *notls,
			Debug:         *debug,
			Session:       *session,
			Status:        *status,
			StatusMessage: *statusMessage,
		}

		talk, err = options.NewClient()

		if err != nil {
			t.Error(err.Error())
		}

		go func() {
			for {
				chat, err := talk.Recv()
				if err != nil {
					log.Fatal(err)
				}
				switch v := chat.(type) {
				case xmpp.Chat:
					if v.Text != TestMessage {
						t.Errorf("Expected received message : %s, actual received message : %s", TestMessage, v.Text)
					}
				case xmpp.Presence:
				}
			}
		}()

		line := strings.TrimRight(*username+" "+TestMessage, "\n")

		tokens := strings.SplitN(line, " ", 2)
		if len(tokens) == 2 {
			talk.Send(xmpp.Chat{Remote: tokens[0], Type: "chat", Text: tokens[1]})
		}

		//waiting for receiving go routine
		time.Sleep(5 * time.Second)
	}
}
