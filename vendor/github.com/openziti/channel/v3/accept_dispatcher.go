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
	"github.com/michaelquigley/pfxlog"
	"time"
)

// An UnderlayAcceptor take an Underlay and generally turns it into a channel for a specific use.
// It can be used when handling multiple channel types on a single listener
type UnderlayAcceptor interface {
	AcceptUnderlay(u Underlay) error
}

type UnderlayDispatcherConfig struct {
	Listener        UnderlayListener
	ConnectTimeout  time.Duration
	Acceptors       map[string]UnderlayAcceptor
	DefaultAcceptor UnderlayAcceptor
}

// An UnderlayDispatcher accept underlays from an underlay listener and hands them off to
// UnderlayAcceptor instances, based on the TypeHeader.
type UnderlayDispatcher struct {
	listener        UnderlayListener
	connectTimeout  time.Duration
	acceptors       map[string]UnderlayAcceptor
	defaultAcceptor UnderlayAcceptor
}

func NewUnderlayDispatcher(config UnderlayDispatcherConfig) *UnderlayDispatcher {
	return &UnderlayDispatcher{
		listener:        config.Listener,
		connectTimeout:  config.ConnectTimeout,
		acceptors:       config.Acceptors,
		defaultAcceptor: config.DefaultAcceptor,
	}
}

func (self *UnderlayDispatcher) Run() {
	log := pfxlog.Logger()
	log.Info("started")
	defer log.Warn("exited")

	for {
		underlay, err := self.listener.Create(self.connectTimeout)
		if err != nil {
			log.WithError(err).Error("error accepting connection")
			if err.Error() == "closed" {
				return
			}
			continue
		}
		chanType, found := underlay.Headers()[TypeHeader]
		var acceptor UnderlayAcceptor

		if !found {
			acceptor = self.defaultAcceptor
		} else {
			acceptor = self.acceptors[string(chanType)]
		}

		closeUnderlay := false
		if acceptor == nil {
			log.Warn("incoming request didn't have type header, and no default acceptor defined. closing connection")
			closeUnderlay = true
		} else if err = acceptor.AcceptUnderlay(underlay); err != nil {
			log.WithError(err).Error("error handling incoming connection, closing connection")
			closeUnderlay = true
		}

		if closeUnderlay {
			if err = underlay.Close(); err != nil {
				log.WithError(err).Info("error closing connection")
			}
		}
	}
}
