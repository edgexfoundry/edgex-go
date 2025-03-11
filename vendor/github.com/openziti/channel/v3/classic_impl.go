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
	"fmt"
	"github.com/openziti/transport/v2"
	"github.com/pkg/errors"
	"net"
	"sync/atomic"
	"time"
)

type classicImpl struct {
	peer         transport.Conn
	id           string
	connectionId string
	headers      map[int32][]byte
	closed       atomic.Bool
	readF        readFunction
	marshalF     marshalFunction
}

func (impl *classicImpl) GetLocalAddr() net.Addr {
	return impl.peer.LocalAddr()
}

func (impl *classicImpl) GetRemoteAddr() net.Addr {
	return impl.peer.RemoteAddr()
}

func (impl *classicImpl) SetWriteTimeout(duration time.Duration) error {
	return impl.peer.SetWriteDeadline(time.Now().Add(duration))
}

func (impl *classicImpl) SetWriteDeadline(deadline time.Time) error {
	return impl.peer.SetWriteDeadline(deadline)
}

func (impl *classicImpl) rxHello() (*Message, error) {
	msg, readF, marshallF, err := readHello(impl.peer)
	impl.readF = readF
	impl.marshalF = marshallF
	return msg, err
}

func (impl *classicImpl) Rx() (*Message, error) {
	if impl.closed.Load() {
		return nil, errors.New("underlay closed")
	}
	return impl.readF(impl.peer)
}

func (impl *classicImpl) Tx(m *Message) error {
	if impl.closed.Load() {
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

func (impl *classicImpl) Id() string {
	return impl.id
}

func (impl *classicImpl) Headers() map[int32][]byte {
	return impl.headers
}

func (impl *classicImpl) LogicalName() string {
	return "classic"
}

func (impl *classicImpl) ConnectionId() string {
	return impl.connectionId
}

func (impl *classicImpl) Certificates() []*x509.Certificate {
	return impl.peer.PeerCertificates()
}

func (impl *classicImpl) Label() string {
	return fmt.Sprintf("u{%s}->i{%s/%s}", impl.LogicalName(), impl.id, impl.ConnectionId())
}

func (impl *classicImpl) Close() error {
	if impl.closed.CompareAndSwap(false, true) {
		return impl.peer.Close()
	}
	return nil
}

func (impl *classicImpl) IsClosed() bool {
	return impl.closed.Load()
}

func (impl *classicImpl) init(id string, connectionId string, headers Headers) {
	impl.id = id
	impl.connectionId = connectionId
	impl.headers = headers
}

func (impl *classicImpl) getPeer() transport.Conn {
	return impl.peer
}

func newClassicImpl(messageStrategy MessageStrategy, peer transport.Conn, version uint32) classicUnderlay {
	readF := ReadV2
	marshalF := MarshalV2

	if version == 3 { // currently only used for testing fallback to a common protocol version
		marshalF = marshalV3
	}

	if messageStrategy != nil && messageStrategy.GetStreamProducer() != nil {
		readF = messageStrategy.GetStreamProducer()
	}

	if messageStrategy != nil && messageStrategy.GetMarshaller() != nil {
		marshalF = messageStrategy.GetMarshaller()
	}

	return &classicImpl{
		peer:     peer,
		readF:    readF,
		marshalF: marshalF,
	}
}
