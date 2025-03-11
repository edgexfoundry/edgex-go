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
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type connImpl struct {
	ws       *websocket.Conn
	leftover []byte
	log      *logrus.Entry
	cfg      *Config
	mu       sync.Mutex
}

func (self *connImpl) Read(b []byte) (int, error) {
	if len(self.leftover) > 0 {
		n := copy(b, self.leftover)
		self.leftover = self.leftover[n:]
		return n, nil
	}

	_, buf, err := self.ws.ReadMessage()
	if err != nil {
		return 0, err
	}

	n := copy(buf, self.leftover)
	self.leftover = buf[n:]
	return n, nil
}

func (self *connImpl) Write(b []byte) (int, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if err := self.ws.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (self *connImpl) Close() error {
	return self.ws.Close()
}

// pinger sends ping messages on an interval for client keep-alive.
func (self *connImpl) pinger() {

	lastResponse := time.Now()

	self.ws.SetPongHandler(func(msg string) error {
		self.log.Debugf("connImpl.pongHandler received websocket Pong: %s", msg)
		lastResponse = time.Now()
		return nil
	})

	ticker := time.NewTicker(self.cfg.PingInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		self.log.Debug("connImpl.pinger sending websocket Ping")
		self.mu.Lock()
		err := self.ws.WriteMessage(websocket.PingMessage, []byte("browzerkeepalive"))
		self.mu.Unlock()
		if err != nil {
			self.log.Warnf("connImpl.pinger: %v", err)
			_ = self.ws.Close()
			return
		}
		if time.Since(lastResponse) > self.cfg.PongTimeout {
			self.log.Errorf("connImpl.pinger PongTimeout exceeded, closing WebSocket")
			self.ws.Close()
			return
		}
	}
}

func (self *connImpl) LocalAddr() net.Addr {
	return self.ws.LocalAddr()
}

func (self *connImpl) RemoteAddr() net.Addr {
	return self.ws.RemoteAddr()
}

func (self *connImpl) SetDeadline(t time.Time) error {
	return self.ws.SetReadDeadline(t)
}

func (self *connImpl) SetReadDeadline(t time.Time) error {
	return self.ws.SetReadDeadline(t)
}

func (self *connImpl) SetWriteDeadline(t time.Time) error {
	return self.ws.SetWriteDeadline(t)
}
