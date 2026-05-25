//go:build js

/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package wss

import (
	"context"
	"crypto/tls"
	"net/url"
	"time"

	"nhooyr.io/websocket"

	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
	transporttls "github.com/openziti/transport/v2/tls"
	log "github.com/sirupsen/logrus"
)

func Dial(name string, u url.URL, i *identity.TokenId, to time.Duration, _ transport.Configuration) (transport.Conn, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute) //cancel //time.Minute)

	log.Debugf("Dialing websocket: %", u.String())
	c, httpResp, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		return nil, err
	}
	log.Debugf("httpResp %v", httpResp)

	conn := websocket.NetConn(ctx, c, websocket.MessageBinary)
	tlsConn := tls.Client(conn, ClientTLSConfig(u, i))

	detail := &transport.ConnectionDetail{
		Address: Type + ":" + u.Host,
		InBound: false,
		Name:    name,
	}
	return transporttls.NewConnection(detail, tlsConn), nil
}

func DialWithLocalBinding(name string, u url.URL, localBinding string, i *identity.TokenId, timeout time.Duration, tcfg transport.Configuration) (transport.Conn, error) {
	return Dial(name, u, i, timeout, tcfg)
}
