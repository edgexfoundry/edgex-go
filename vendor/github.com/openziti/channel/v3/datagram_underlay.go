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
	"bytes"
	"crypto/x509"
	"fmt"
	"github.com/openziti/transport/v2"
	"net"
	"sync/atomic"
	"time"
)

type readPacketFunction func(buf []byte) (*Message, error)

type DatagramUnderlay struct {
	id           string
	connectionId string
	headers      map[int32][]byte
	peer         transport.Conn
	closed       atomic.Bool
	readF        readPacketFunction
	marshalF     marshalFunction
}

func newDatagramUnderlay(messageStrategy MessageStrategy, peer transport.Conn, version uint32) classicUnderlay {
	readF := func(buf []byte) (*Message, error) {
		reader := bytes.NewBuffer(buf)
		return ReadV2(reader)
	}

	marshalF := MarshalV2

	if version == 3 { // currently only used for testing fallback to a common protocol version
		marshalF = marshalV3
	}

	if messageStrategy != nil && messageStrategy.GetPacketProducer() != nil {
		readF = messageStrategy.GetPacketProducer()
	}

	if messageStrategy != nil && messageStrategy.GetMarshaller() != nil {
		marshalF = messageStrategy.GetMarshaller()
	}

	return &DatagramUnderlay{
		peer:     peer,
		readF:    readF,
		marshalF: marshalF,
	}
}

func (self *DatagramUnderlay) GetLocalAddr() net.Addr {
	return self.peer.LocalAddr()
}

func (self *DatagramUnderlay) GetRemoteAddr() net.Addr {
	return self.peer.RemoteAddr()
}

func (self *DatagramUnderlay) Rx() (*Message, error) {
	buf := make([]byte, 2000)
	n, err := self.peer.Read(buf)
	if err != nil {
		return nil, err
	}

	buf = buf[:n]
	return self.readF(buf)
}

func (self *DatagramUnderlay) Tx(m *Message) error {
	data, err := self.marshalF(m)
	if err != nil {
		return err
	}
	_, err = self.peer.Write(data)
	return err
}

func (self *DatagramUnderlay) Id() string {
	return self.id
}

func (self *DatagramUnderlay) LogicalName() string {
	return "datagram"
}

func (self *DatagramUnderlay) ConnectionId() string {
	return self.connectionId
}

func (self *DatagramUnderlay) Certificates() []*x509.Certificate {
	return self.peer.PeerCertificates()
}

func (self *DatagramUnderlay) Label() string {
	return fmt.Sprintf("u{%s}->i{%s}", self.LogicalName(), self.ConnectionId())
}

func (self *DatagramUnderlay) Close() error {
	if self.closed.CompareAndSwap(false, true) {
		return self.peer.Close()
	}
	return nil
}

func (self *DatagramUnderlay) IsClosed() bool {
	return self.closed.Load()
}

func (self *DatagramUnderlay) Headers() map[int32][]byte {
	return self.headers
}

func (self *DatagramUnderlay) SetWriteTimeout(duration time.Duration) error {
	return self.peer.SetWriteDeadline(time.Now().Add(duration))
}

func (self *DatagramUnderlay) SetWriteDeadline(deadline time.Time) error {
	return self.peer.SetWriteDeadline(deadline)
}

func (impl *DatagramUnderlay) init(id string, connectionId string, headers Headers) {
	impl.id = id
	impl.connectionId = connectionId
	impl.headers = headers
}

func (impl *DatagramUnderlay) getPeer() transport.Conn {
	return impl.peer
}

func (self *DatagramUnderlay) rxHello() (*Message, error) {
	return self.Rx()
}

type DatagramMessageStrategy PacketMessageProducer

func (self DatagramMessageStrategy) GetMarshaller() MessageMarshaller {
	return MarshalV2WithRaw
}

func (self DatagramMessageStrategy) GetStreamProducer() StreamMessageProducer {
	return nil
}

func (self DatagramMessageStrategy) GetPacketProducer() PacketMessageProducer {
	return PacketMessageProducer(self)
}
