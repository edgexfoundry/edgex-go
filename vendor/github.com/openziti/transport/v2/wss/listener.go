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
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/openziti/identity"
	transporttls "github.com/openziti/transport/v2/tls"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/transport/v2"
	"github.com/sirupsen/logrus"
)

var (
	browZerRuntimeSdkSuites = []uint16{
		//vv WASM-based TLS1.3 suites
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_AES_128_GCM_SHA256,
		//^^
	}
)

type wssListener struct {
	log      *logrus.Entry
	acceptF  func(transport.Conn)
	cfg      *Config
	ctr      int64
	upgrader websocket.Upgrader
}

/**
 *	Accept acceptF HTTP connection, and upgrade it to a websocket suitable for communication between ziti-browzer-runtime and Ziti Edge Router
 */
func (listener *wssListener) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	log := listener.log
	log.Info("entered")

	c, err := listener.upgrader.Upgrade(w, r, nil) // upgrade from HTTP to binary socket

	if err != nil {
		log.WithError(err).Error("websocket upgrade failed. Failure not recoverable.")
	} else {

		var zero time.Time
		_ = c.SetReadDeadline(zero)

		listener.ctr++

		cfg := listener.cfg.Identity.ServerTLSConfig()
		cfg.ClientAuth = tls.RequireAndVerifyClientCert

		// This is technically not correct but will help get work moving forward.
		// Instead of using ClientCAs we should rely on VerifyPeerCertificate
		// or VerifyConnection similar to how the controller does it
		cfg.ClientCAs = cfg.RootCAs
		cfg.CipherSuites = append(cfg.CipherSuites, browZerRuntimeSdkSuites...)

		connWrapper := &connImpl{
			ws:  c,
			log: log,
			cfg: listener.cfg,
		}

		tlsConn := tls.Server(connWrapper, cfg)
		if err = tlsConn.Handshake(); err != nil {
			log.WithError(err).Error("unable to establish tls over websocket")
			_ = c.Close()
			return
		}

		detail := &transport.ConnectionDetail{
			Address: Type + ":" + c.NetConn().RemoteAddr().String(),
			InBound: true,
			Name:    Type,
		}

		connection := transporttls.NewConnection(detail, tlsConn)
		listener.acceptF(connection) // pass the Websocket to the goroutine that will validate the HELLO handshake

		// keep the Websocket alive via ping/pong control-frame msgs
		// so it doesn't close unnecessarily thus causing ZBR to encounter
		// unnecessary 'channel unavailable' conditions thus causing too
		// frequent Page reboots on the client-side
		go connWrapper.pinger()
	}
}

func Listen(bindAddress string, name string, i *identity.TokenId, acceptF func(transport.Conn), tcfg transport.Configuration) (io.Closer, error) {
	log := pfxlog.ContextLogger(name + "/" + Type + ":" + bindAddress)

	cfg := NewDefaultConfig()
	cfg.Identity = i

	if tcfg != nil {
		if err := cfg.Load(tcfg); err != nil {
			return nil, fmt.Errorf("error loading configuration: %w", err)
		}
	}
	logrus.Info(cfg.Dump("ws.Config"))

	log.Infof("starting HTTP (websocket) server at bindAddress [%s]", bindAddress)

	listener := &wssListener{
		log:     log.Entry,
		acceptF: acceptF,
		cfg:     cfg,
		ctr:     0,
	}

	// Set up the HTTP -> Websocket upgrader options with a local upgrader to avoid races
	listener.upgrader = websocket.Upgrader{
		HandshakeTimeout:  cfg.HandshakeTimeout,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		EnableCompression: cfg.EnableCompression,
		CheckOrigin:       func(r *http.Request) bool { return true },
	}

	router := mux.NewRouter()
	router.HandleFunc("/ws", listener.handleWebsocket).Methods("GET")

	tlsConfig := cfg.Identity.ServerTLSConfig()
	tlsConfig.ClientAuth = tls.NoClientCert
	tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2", "http/1.1")

	httpServer := &http.Server{
		Addr:         bindAddress,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      router,
		TLSConfig:    tlsConfig,
	}

	nl, err := transporttls.ListenTLS(bindAddress, "wss", tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("listen TLS error: %w", err)
	}

	go func() {
		if err := httpServer.Serve(nl); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Entry.WithError(err).Error("HTTP server failed")
		}
	}()

	return httpServer, nil
}
