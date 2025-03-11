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
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
	"github.com/pkg/errors"
	"io"
	"net"
	"sync/atomic"
	"time"
)

func (impl *reconnectingImpl) Rx() (*Message, error) {
	log := pfxlog.ContextLogger(impl.Label())

	connected := true
	for !impl.closed.Load() {
		if connected {
			m, err := impl.rx()
			if err != nil {
				if closeErr := impl.peer.Close(); closeErr != nil {
					log.WithError(closeErr).Error("error closing peer after rx error")
				}
				log.WithError(err).Error("rx error. closed peer and starting reconnection process")
				connected = false
			} else {
				return m, nil
			}
		} else {
			if err := impl.reconnectionHandler.Reconnect(impl); err != nil {
				log.Errorf("reconnection failed (%s)", err)
				return nil, fmt.Errorf("reconnection failed (%s)", err)
			} else {
				log.Info("reconnected")
				connected = true
			}
		}
	}
	return nil, io.EOF
}

func (impl *reconnectingImpl) Tx(m *Message) error {
	log := pfxlog.ContextLogger(impl.Label())

	done := false
	connected := true
	for !done && !impl.closed.Load() {
		if connected {
			if err := impl.tx(m); err != nil {
				log.Errorf("tx error (%s). starting reconnection process", err)
				connected = false
			} else {
				done = true
			}
		} else {
			if err := impl.reconnectionHandler.Reconnect(impl); err != nil {
				log.Errorf("reconnection failed (%s)", err)
				return fmt.Errorf("reconnection failed (%s)", err)

			} else {
				log.Info("reconnected")
				connected = true
			}
		}
	}
	return nil
}

func (impl *reconnectingImpl) Id() string {
	return impl.id.Token
}

func (impl *reconnectingImpl) Headers() map[int32][]byte {
	return impl.headers.Load()
}

func (impl *reconnectingImpl) LogicalName() string {
	return "reconnecting"
}

func (impl *reconnectingImpl) ConnectionId() string {
	return impl.connectionId
}

func (impl *reconnectingImpl) Certificates() []*x509.Certificate {
	return impl.peer.PeerCertificates()
}

func (impl *reconnectingImpl) Label() string {
	return fmt.Sprintf("u{%s}->i{%s/%s}", impl.LogicalName(), impl.id.Token, impl.ConnectionId())
}

func (impl *reconnectingImpl) Close() error {
	if impl.closed.CompareAndSwap(false, true) {
		return impl.peer.Close()
	}
	return nil
}

func (impl *reconnectingImpl) IsClosed() bool {
	return impl.closed.Load()
}

func newReconnectingImpl(peer transport.Conn, reconnectionHandler reconnectionHandler, timeout time.Duration) *reconnectingImpl {
	id := &identity.TokenId{Token: "unknown"}
	if certs := peer.PeerCertificates(); len(certs) > 0 {
		id = &identity.TokenId{Token: certs[0].Subject.CommonName}
	}

	return &reconnectingImpl{
		id:                  id,
		peer:                peer,
		reconnectionHandler: reconnectionHandler,
		readF:               ReadV2,
		marshalF:            MarshalV2,
		timeout:             timeout,
	}
}

func (impl *reconnectingImpl) setProtocolVersion(version uint32) {
	if version == 2 {
		impl.readF = ReadV2
		impl.marshalF = MarshalV2
	} else {
		pfxlog.Logger().Warnf("asked to set unsupported protocol version %v", version)
	}
}

func (impl *reconnectingImpl) rx() (*Message, error) {
	return impl.readF(impl.peer)
}

func (impl *reconnectingImpl) tx(m *Message) error {
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

// pingInstance currently does a single-sided (unverified) ping to see if the peer connection is functional.
func (impl *reconnectingImpl) pingInstance() error {
	log := pfxlog.ContextLogger(impl.Label())
	defer log.Info("exiting")
	log.Info("starting")

	ping := NewMessage(reconnectingPingContentType, nil)
	if err := impl.tx(ping); err != nil {
		return err
	}

	return nil
}

func (impl *reconnectingImpl) Disconnect() error {
	if dialer, ok := impl.reconnectionHandler.(*reconnectingDialer); ok {
		if impl.disconnected.CompareAndSwap(false, true) {
			dialer.reconnectLock.Lock()
			return impl.peer.Close()
		} else {
			return errors.New("already marked disconnected")
		}
	} else {
		return errors.New("unexpected reconnect handler implementation")
	}
}

func (impl *reconnectingImpl) Reconnect() error {
	if dialer, ok := impl.reconnectionHandler.(*reconnectingDialer); ok {
		if impl.disconnected.CompareAndSwap(true, false) {
			dialer.reconnectLock.Unlock()
			return nil
		} else {
			return errors.New("cannot reconnect, not disconnected")
		}
	} else {
		return errors.New("unexpected reconnect handler implementation")
	}
}

func (impl *reconnectingImpl) SetWriteTimeout(duration time.Duration) error {
	return impl.peer.SetWriteDeadline(time.Now().Add(duration))
}

func (impl *reconnectingImpl) SetWriteDeadline(deadline time.Time) error {
	return impl.peer.SetWriteDeadline(deadline)
}

func (impl *reconnectingImpl) IsConnected() bool {
	return !impl.reconnecting.Load() && !impl.disconnected.Load()
}

type reconnectingImpl struct {
	peer                transport.Conn
	id                  *identity.TokenId
	connectionId        string
	headers             concurrenz.AtomicValue[map[int32][]byte]
	reconnectionHandler reconnectionHandler
	closed              atomic.Bool
	readF               readFunction
	marshalF            marshalFunction
	disconnected        atomic.Bool
	reconnecting        atomic.Bool
	timeout             time.Duration
}

func (impl *reconnectingImpl) GetLocalAddr() net.Addr {
	return impl.peer.LocalAddr()
}

func (impl *reconnectingImpl) GetRemoteAddr() net.Addr {
	return impl.peer.RemoteAddr()
}

type reconnectionHandler interface {
	Reconnect(impl *reconnectingImpl) error
}

const reconnectingPingContentType = -33
