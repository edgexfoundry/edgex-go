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

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/openziti/transport/v2"
)

func NewConnection(detail *transport.ConnectionDetail, conn *tls.Conn) *Connection {
	return &Connection{
		detail: detail,
		Conn:   conn,
	}
}

type Connection struct {
	detail *transport.ConnectionDetail
	*tls.Conn
}

func (self *Connection) Detail() *transport.ConnectionDetail {
	return self.detail
}

func (self *Connection) PeerCertificates() []*x509.Certificate {
	_ = self.Handshake()
	return self.ConnectionState().PeerCertificates
}

func (self *Connection) Protocol() string {
	return self.ConnectionState().NegotiatedProtocol
}
