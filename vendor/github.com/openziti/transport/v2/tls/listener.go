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
	"context"
	"crypto/tls"
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/foundation/v2/rate"
	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// same as golang Dial default
	keepAlive = 15 * time.Second
)

var noProtocol = ""
var handlerKey = struct{}{}

var handshakeTimeout concurrenz.AtomicValue[time.Duration]

func SetSharedListenerHandshakeTimeout(timeout time.Duration) {
	handshakeTimeout.Store(timeout)
}

var rateLimiterAtomic concurrenz.AtomicValue[*rate.AdaptiveRateLimitTracker]

func SetSharedListenerRateLimiter(limiter rate.AdaptiveRateLimitTracker) {
	rateLimiterAtomic.Store(&limiter)
}

func init() {
	var limiter rate.AdaptiveRateLimitTracker = rate.NoOpAdaptiveRateLimitTracker{}
	rateLimiterAtomic.Store(&limiter)
}

func Listen(bindAddress, name string, i *identity.TokenId, acceptF func(transport.Conn), protocols ...string) (io.Closer, error) {
	log := pfxlog.ContextLogger(name + "/" + Type + ":" + bindAddress).Entry

	config := i.ServerTLSConfig().Clone()
	if len(protocols) > 0 {
		config.NextProtos = append(config.NextProtos, protocols...)
	}
	result := &protocolHandler{
		name:    name,
		tls:     config,
		acceptF: acceptF,
	}

	err := registerWithSharedListener(bindAddress, result)
	if err != nil {
		log.WithError(err).Error("failed to register with shared listener")
		return nil, err
	}

	return result, nil
}

type tlsListener struct {
	connCh  chan *Connection
	handler *protocolHandler
	closed  atomic.Bool
}

func (self *tlsListener) Accept() (net.Conn, error) {
	conn := <-self.connCh
	if conn == nil {
		return nil, net.ErrClosed
	}
	return conn.Conn, nil
}

func (self *tlsListener) Close() error {
	var err error
	if self.closed.CompareAndSwap(false, true) {
		err = self.handler.Close()
		close(self.connCh)
	}
	return err
}

func (self *tlsListener) Addr() net.Addr {
	return self.handler.listener.sock.Addr()
}

func (self *tlsListener) tlsAccept(conn transport.Conn) {
	c := conn.(*Connection)
	self.connCh <- c
}

// ListenTLS returns net.Listener that is attached to shared listener with protocols (ALPN)
// specified by config.NextProtos
// It can be used in http.Server or other standard components
func ListenTLS(bindAddress, name string, config *tls.Config) (net.Listener, error) {
	log := pfxlog.ContextLogger(name + "/" + Type + ":" + bindAddress).Entry

	l := &tlsListener{}

	handler := &protocolHandler{
		name:    name,
		tls:     config,
		acceptF: l.tlsAccept,
	}

	err := registerWithSharedListener(bindAddress, handler)
	if err != nil {
		log.WithError(err).Error("failed to register with shared listener")
		return nil, err
	}

	l.handler = handler
	l.connCh = make(chan *Connection, 16)

	return l, nil
}

type protocolHandler struct {
	name     string
	listener *sharedListener
	tls      *tls.Config
	acceptF  func(conn transport.Conn)
	closed   atomic.Bool
}

func (self *protocolHandler) Close() error {
	if self.closed.CompareAndSwap(false, true) {
		self.listener.remove(self)
		return nil
	}
	return nil
}

var sharedListeners sync.Map

func registerWithSharedListener(bindAddress string, acc *protocolHandler) error {
	sl := &sharedListener{
		address: bindAddress,
	}
	el, found := sharedListeners.LoadOrStore(bindAddress, sl)
	sl = el.(*sharedListener)

	if !found {
		sl.log = pfxlog.ContextLogger(Type + ":" + bindAddress).Entry

		sl.tlsCfg = &tls.Config{
			GetConfigForClient: sl.getConfig,
		}

		sl.ctx, sl.done = context.WithCancel(context.Background())
		sl.handlers = make(map[string]*protocolHandler)
		sock, err := tls.Listen("tcp", bindAddress, sl.tlsCfg)
		if err != nil {
			sharedListeners.Delete(bindAddress)
			return err
		}
		sl.sock = sock
		go sl.runAccept()
	}

	protos := acc.tls.NextProtos
	if protos == nil {
		protos = append(protos, "")
	}

	sl.mtx.Lock()
	defer sl.mtx.Unlock()

	// check for conflict
	for _, proto := range protos {
		if _, exists := sl.handlers[proto]; exists {
			return fmt.Errorf("handler for protocol[%s] already exists", proto)
		}
	}

	acc.listener = sl
	for _, proto := range protos {
		sl.handlers[proto] = acc
	}

	return nil
}

type sharedListener struct {
	log      logrus.FieldLogger
	address  string
	tlsCfg   *tls.Config
	mtx      sync.RWMutex
	handlers map[string]*protocolHandler // proto -> protocolHandler
	ctx      context.Context
	done     context.CancelFunc
	sock     net.Listener
}

func (self *sharedListener) processConn(conn *tls.Conn) {
	log := self.log.WithField("remote", conn.RemoteAddr().String())

	if tcpConn, ok := conn.NetConn().(*net.TCPConn); ok {
		_ = tcpConn.SetNoDelay(true)
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(keepAlive)
	}

	timeout := handshakeTimeout.Load()
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	rateLimiter := *rateLimiterAtomic.Load()

	// sharedListener.getConfig will select the right handler during handshake based on ClientHelloInfo
	// no need to do another look up here
	var handler *protocolHandler
	hsCtx, cancelF := context.WithTimeout(context.WithValue(self.ctx, handlerKey, &handler), timeout)
	defer cancelF()

	handshakeF := func(control rate.RateLimitControl) error {
		err := conn.HandshakeContext(hsCtx)
		if err != nil {
			if io.EOF == err {
				control.Backoff()
			} else {
				control.Failed()
			}
			return err
		}
		control.Success()
		return nil
	}

	err := rateLimiter.RunRateLimitedF(fmt.Sprintf("tls handlshake from %s", conn.RemoteAddr().String()), handshakeF)

	if err != nil {
		log.WithError(err).Error("handshake failed")
		_ = conn.Close()
		return
	}

	proto := conn.ConnectionState().NegotiatedProtocol
	log.WithField("client", conn.RemoteAddr()).Debug("selected protocol = '", proto, "'")

	connection := &Connection{
		detail: &transport.ConnectionDetail{
			Address: Type + ":" + conn.RemoteAddr().String(),
			InBound: true,
			Name:    handler.name,
		},
		Conn: conn,
	}
	handler.acceptF(connection)
}

func (self *sharedListener) runAccept() {
	log := self.log
	defer log.Info("exited")
	for {
		c, err := self.sock.Accept()
		if err != nil {
			if self.ctx.Err() != nil {
				log.WithError(err).Info("listener closed, exiting")
				return
			}
			log.WithError(err).Error("accept failed, exiting")
			return
		}

		conn := c.(*tls.Conn)

		go self.processConn(conn)
	}
}

func (self *sharedListener) getConfig(info *tls.ClientHelloInfo) (*tls.Config, error) {
	log := self.log.WithField("client", info.Conn.RemoteAddr())

	protos := info.SupportedProtos
	log.Debug("client requesting protocols = ", protos)

	ctx := info.Context()
	handlerOut := ctx.Value(handlerKey).(**protocolHandler)

	self.mtx.RLock()
	defer self.mtx.RUnlock()

	var handler *protocolHandler
	var proto string
	if protos == nil && len(self.handlers) == 1 {
		log.Debugf("using single protocol as default")
		for p, h := range self.handlers {
			proto, handler = p, h
		}
	} else {
		if protos == nil {
			protos = append(protos, noProtocol)
		}

		for _, p := range protos {
			h, found := self.handlers[p]
			if found {
				log.Debugf("found handler for proto[%s]", proto)
				handler = h
				proto = p
			}
		}
	}

	if handler != nil {
		*handlerOut = handler
		cfg := handler.tls
		if cfg.GetConfigForClient != nil {
			c, _ := cfg.GetConfigForClient(info)
			if c != nil {
				cfg = c
			}
		}
		cfg = cfg.Clone()
		cfg.NextProtos = []string{proto}
		return cfg, nil
	}

	return nil, fmt.Errorf("not handler for requested protocols %+v", protos)
}

func (self *sharedListener) remove(h *protocolHandler) {
	self.log.WithField("name", h.name).Debug("removing handler")

	protos := h.tls.NextProtos
	if protos == nil {
		protos = append(protos, noProtocol)
	}

	for _, p := range protos {
		delete(self.handlers, p)
	}

	self.mtx.Lock()
	defer self.mtx.Unlock()

	if len(self.handlers) == 0 {
		self.log.Debug("no handlers left. stopping")
		sharedListeners.Delete(self.address)
		self.done()
		_ = self.sock.Close()
	}
}
