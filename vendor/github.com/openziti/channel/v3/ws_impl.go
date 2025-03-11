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

func (impl *wsImpl) SetWriteTimeout(duration time.Duration) error {
	return impl.peer.SetWriteDeadline(time.Now().Add(duration))
}

func (self *wsImpl) SetWriteDeadline(deadline time.Time) error {
	return self.peer.SetWriteDeadline(deadline)
}

func (impl *wsImpl) Rx() (*Message, error) {
	if impl.closed {
		return nil, errors.New("underlay closed")
	}
	return impl.readF(impl.peer)
}

func (impl *wsImpl) Tx(m *Message) error {
	if impl.closed {
		return errors.New("underlay closed")
	}

	data, err := impl.marshalF(m)
	if err != nil {
		return err
	}

	_, err = impl.peer.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (impl *wsImpl) Id() string {
	return impl.id.Token
}

func (impl *wsImpl) Headers() map[int32][]byte {
	return impl.headers
}

func (impl *wsImpl) LogicalName() string {
	return "ws"
}

func (impl *wsImpl) ConnectionId() string {
	return impl.connectionId
}

func (impl *wsImpl) Certificates() []*x509.Certificate {
	return impl.peer.PeerCertificates()
}

func (impl *wsImpl) Label() string {
	return fmt.Sprintf("u{%s}->i{%s}", impl.LogicalName(), impl.ConnectionId())
}

func (impl *wsImpl) Close() error {
	impl.closeLock.Lock()
	defer impl.closeLock.Unlock()

	if !impl.closed {
		impl.closed = true
		return impl.peer.Close()
	}
	return nil
}

func (impl *wsImpl) IsClosed() bool {
	return impl.closed
}

func newWSImpl(peer transport.Conn, version uint32) *wsImpl {
	readF := ReadV2
	marshalF := MarshalV2

	if version == 2 {
		readF = ReadV2
		marshalF = MarshalV2
	} else if version == 3 { // currently only used for testing fallback to a common protocol version
		readF = ReadV2
		marshalF = marshalV3
	}

	return &wsImpl{
		peer:     peer,
		readF:    readF,
		marshalF: marshalF,
	}
}
