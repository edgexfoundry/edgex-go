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
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/goroutines"
	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"sync/atomic"
	"time"
)

type classicListener struct {
	identity        *identity.TokenId
	endpoint        transport.Address
	socket          io.Closer
	close           chan struct{}
	handlers        []ConnectionHandler
	acceptF         func(underlay Underlay)
	created         chan Underlay
	connectOptions  ConnectOptions
	tcfg            transport.Configuration
	headers         map[int32][]byte
	headersF        func() map[int32][]byte
	closed          atomic.Bool
	listenerPool    goroutines.Pool
	messageStrategy MessageStrategy
	underlayFactory func(messageStrategy MessageStrategy, peer transport.Conn, version uint32) classicUnderlay
}

func DefaultListenerConfig() ListenerConfig {
	return ListenerConfig{
		ConnectOptions: DefaultConnectOptions(),
	}
}

type ListenerConfig struct {
	ConnectOptions
	Headers            map[int32][]byte
	HeadersF           func() map[int32][]byte
	TransportConfig    transport.Configuration
	PoolConfigurator   func(config *goroutines.PoolConfig)
	ConnectionHandlers []ConnectionHandler
	MessageStrategy    MessageStrategy
}

func newClassicListener(identity *identity.TokenId, endpoint transport.Address, config ListenerConfig) *classicListener {
	closeNotify := make(chan struct{})

	poolConfig := goroutines.PoolConfig{
		QueueSize:  uint32(config.ConnectOptions.MaxQueuedConnects),
		MinWorkers: 1,
		MaxWorkers: uint32(config.ConnectOptions.MaxOutstandingConnects),
		IdleTime:   10 * time.Second,
		PanicHandler: func(err interface{}) {
			pfxlog.Logger().
				WithField("id", identity.Token).
				WithField("endpoint", endpoint.String()).
				WithField(logrus.ErrorKey, err).Error("panic during channel accept")
		},
	}

	if config.PoolConfigurator != nil {
		config.PoolConfigurator(&poolConfig)
	}

	poolConfig.CloseNotify = closeNotify

	pool, err := goroutines.NewPool(poolConfig)
	if err != nil {
		logrus.WithError(err).Error("failed to initial channel listener pool")
		panic(err)
	}

	underlayFactory := newClassicImpl
	if endpoint.Type() == "dtls" {
		underlayFactory = newDatagramUnderlay
	}

	return &classicListener{
		identity:        identity,
		endpoint:        endpoint,
		socket:          nil,
		close:           closeNotify,
		handlers:        config.ConnectionHandlers,
		acceptF:         nil,
		created:         nil,
		connectOptions:  config.ConnectOptions,
		tcfg:            config.TransportConfig,
		headers:         config.Headers,
		headersF:        config.HeadersF,
		closed:          atomic.Bool{},
		listenerPool:    pool,
		messageStrategy: config.MessageStrategy,
		underlayFactory: underlayFactory,
	}
}

func NewClassicListenerF(identity *identity.TokenId, endpoint transport.Address, config ListenerConfig, f func(underlay Underlay)) (io.Closer, error) {
	listener := newClassicListener(identity, endpoint, config)
	listener.acceptF = f
	if err := listener.Listen(); err != nil {
		return nil, err
	}
	return listener, nil
}

func NewClassicListener(identity *identity.TokenId, endpoint transport.Address, config ListenerConfig) UnderlayListener {
	listener := newClassicListener(identity, endpoint, config)
	listener.created = make(chan Underlay)
	listener.acceptF = func(underlay Underlay) {
		select {
		case listener.created <- underlay:
		case <-listener.close:
			pfxlog.Logger().WithField("underlay", underlay.Label()).Info("channel closed, can't notify of new connection")
			return
		}
	}
	return listener
}

func (self *classicListener) Listen(handlers ...ConnectionHandler) error {
	self.handlers = append(self.handlers, handlers...)
	socket, err := self.endpoint.Listen("classic", self.identity, self.acceptConnection, self.tcfg)
	if err != nil {
		return err
	}
	self.socket = socket
	return nil
}

func (self *classicListener) Close() error {
	if self.closed.CompareAndSwap(false, true) {
		close(self.close)
		if socket := self.socket; socket != nil {
			if err := socket.Close(); err != nil {
				return err
			}
		}
		self.socket = nil
	}
	return nil
}

func (self *classicListener) Create(_ time.Duration) (Underlay, error) {
	if self.created == nil {
		return nil, errors.New("this listener was not set up for Create to be called, programming error")
	}

	select {
	case impl := <-self.created:
		if impl != nil {
			return impl, nil
		}
	case <-self.close:
	}
	return nil, ListenerClosedError
}

func (self *classicListener) acceptConnection(peer transport.Conn) {
	log := pfxlog.ContextLogger(self.endpoint.String())
	err := self.listenerPool.Queue(func() {
		impl := self.underlayFactory(self.messageStrategy, peer, 2)

		connectionId, err := NextConnectionId()
		if err != nil {
			_ = peer.Close()
			log.Errorf("error getting connection id for [%s] (%v)", peer.Detail().Address, err)
			return
		}

		if err = peer.SetDeadline(time.Now().Add(self.connectOptions.ConnectTimeout)); err != nil {
			log.Errorf("could not set connection deadline for [%s] (%v)", peer.Detail().Address, err)
			_ = peer.Close()
			return
		}

		defer func() {
			if err = peer.SetDeadline(time.Time{}); err != nil {
				log.Errorf("could not clear connection deadline for [%s] (%v)", peer.Detail().Address, err)
				_ = peer.Close()
				return
			}
		}()

		request, hello, err := self.receiveHello(impl)
		if err != nil {
			_ = peer.Close()
			log.Errorf("error receiving hello from [%s] (%v)", peer.Detail().Address, err)
			return
		}

		for _, h := range self.handlers {
			if err = h.HandleConnection(hello, peer.PeerCertificates()); err != nil {
				break
			}
		}

		if err != nil {
			log.Errorf("connection handler error for [%s] (%v)", peer.Detail().Address, err)
			_ = peer.Close()
			return
		}

		impl.init(hello.IdToken, connectionId, hello.Headers)

		if err = self.ackHello(impl, request, true, ""); err == nil {
			self.acceptF(impl)
		} else {
			log.Errorf("error acknowledging hello for [%s] (%v)", peer.Detail().Address, err)
			_ = peer.Close()
		}
	})
	if err != nil {
		log.WithError(err).Error("failed to queue connection accept")
	}
}

func (self *classicListener) receiveHello(impl classicUnderlay) (*Message, *Hello, error) {
	log := pfxlog.ContextLogger(impl.Label())
	log.Debug("started")
	defer log.Debug("exited")

	request, err := impl.rxHello()
	if err != nil {
		if errors.Is(err, BadMagicNumberError) {
			WriteUnknownVersionResponse(impl.getPeer())
		}
		_ = impl.Close()
		return nil, nil, fmt.Errorf("receive error (%s)", err)
	}
	if request.ContentType != ContentTypeHelloType {
		_ = impl.Close()
		return nil, nil, fmt.Errorf("unexpected content type [%d]", request.ContentType)
	}
	hello := UnmarshalHello(request)
	return request, hello, nil
}

func (self *classicListener) ackHello(impl classicUnderlay, request *Message, success bool, message string) error {
	response := NewResult(success, message)

	for key, val := range self.headers {
		response.Headers[key] = val
	}

	if self.headersF != nil {
		for key, val := range self.headersF() {
			response.Headers[key] = val
		}
	}

	response.PutStringHeader(ConnectionIdHeader, impl.ConnectionId())
	if self.identity != nil {
		response.PutStringHeader(IdHeader, self.identity.Token)
	}
	response.sequence = HelloSequence

	response.ReplyTo(request)
	return impl.Tx(response)
}
