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

package channel

import (
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
	"sync"
	"time"
)

type wsImpl struct {
	peer         transport.Conn
	id           *identity.TokenId
	connectionId string
	headers      map[int32][]byte
	closeLock    sync.Mutex
	closed       bool
	readF        readFunction
	marshalF     marshalFunction
}

func (self *wsImpl) SetWriteTimeout(duration time.Duration) error {
	return self.peer.SetWriteDeadline(time.Now().Add(duration))
}

func (self *wsImpl) SetWriteDeadline(deadline time.Time) error {
	return self.peer.SetWriteDeadline(deadline)
}

func (self *wsImpl) Rx() (*Message, error) {
	if self.closed {
		return nil, errors.New("underlay closed")
	}
	return self.readF(self.peer)
}

func (self *wsImpl) Tx(m *Message) error {
	if self.closed {
		return errors.New("underlay closed")
	}

	data, err := self.marshalF(m)
	if err != nil {
		return err
	}

	_, err = self.peer.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (self *wsImpl) Id() string {
	return self.id.Token
}

func (self *wsImpl) Headers() map[int32][]byte {
	return self.headers
}

func (self *wsImpl) LogicalName() string {
	return "ws"
}

func (self *wsImpl) ConnectionId() string {
	return self.connectionId
}

func (self *wsImpl) Certificates() []*x509.Certificate {
	return self.peer.PeerCertificates()
}

func (self *wsImpl) Label() string {
	return fmt.Sprintf("u{%s}->i{%s}", self.LogicalName(), self.ConnectionId())
}

func (self *wsImpl) Close() error {
	self.closeLock.Lock()
	defer self.closeLock.Unlock()

	if !self.closed {
		self.closed = true
		return self.peer.Close()
	}
	return nil
}

func (self *wsImpl) IsClosed() bool {
	return self.closed
}

func newWSImpl(peer transport.Conn, version uint32) *wsImpl {
	readF := ReadV2
	marshalF := MarshalV2

	switch version {
	case 2:
		readF = ReadV2
		marshalF = MarshalV2
	case 3:
		readF = ReadV2
		marshalF = marshalV3
	}

	return &wsImpl{
		peer:     peer,
		readF:    readF,
		marshalF: marshalF,
	}
}
