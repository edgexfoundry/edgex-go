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
	"errors"
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
	"time"
)

type classicDialer struct {
	identity        *identity.TokenId
	endpoint        transport.Address
	localBinding    string
	headers         map[int32][]byte
	underlayFactory func(messageStrategy MessageStrategy, peer transport.Conn, version uint32) classicUnderlay
	messageStrategy MessageStrategy
	transportConfig transport.Configuration
}

type DialerConfig struct {
	Identity        *identity.TokenId
	Endpoint        transport.Address
	LocalBinding    string
	Headers         map[int32][]byte
	MessageStrategy MessageStrategy
	TransportConfig transport.Configuration
}

func NewClassicDialer(cfg DialerConfig) DialUnderlayFactory {
	result := &classicDialer{
		identity:        cfg.Identity,
		endpoint:        cfg.Endpoint,
		localBinding:    cfg.LocalBinding,
		headers:         cfg.Headers,
		messageStrategy: cfg.MessageStrategy,
		transportConfig: cfg.TransportConfig,
	}

	if cfg.Endpoint.Type() == "dtls" {
		result.underlayFactory = newDatagramUnderlay
	} else {
		result.underlayFactory = newClassicImpl
	}

	return result
}
func (self *classicDialer) Create(timeout time.Duration) (Underlay, error) {
	return self.CreateWithHeaders(timeout, nil)
}

func (self *classicDialer) CreateWithHeaders(timeout time.Duration, headers map[int32][]byte) (Underlay, error) {
	log := pfxlog.ContextLogger(self.endpoint.String())
	log.Debug("started")
	defer log.Debug("exited")

	if timeout == 0 {
		timeout = 15 * time.Second
	}

	deadline := time.Now().Add(timeout)

	version := uint32(2)
	tryCount := 0

	log.Debugf("Attempting to dial with bind: %s", self.localBinding)

	for time.Now().Before(deadline) {
		peer, err := self.endpoint.DialWithLocalBinding("classic", self.localBinding, self.identity, timeout, self.transportConfig)
		if err != nil {
			return nil, err
		}

		underlay := self.underlayFactory(self.messageStrategy, peer, version)
		if err = self.sendHello(underlay, deadline, headers); err != nil {
			if tryCount > 0 {
				return nil, err
			} else {
				log.WithError(err).Warnf("error initiating channel with hello")
			}
			tryCount++
			version, _ = GetRetryVersion(err)
			log.Warnf("Retrying dial with protocol version %v", version)
			continue
		}
		return underlay, nil
	}
	return nil, errors.New("timeout waiting for dial")
}

func (self *classicDialer) sendHello(underlay classicUnderlay, deadline time.Time, headers map[int32][]byte) error {
	log := pfxlog.ContextLogger(underlay.Label())
	defer log.Debug("exited")
	log.Debug("started")

	peer := underlay.getPeer()

	if err := peer.SetDeadline(deadline); err != nil {
		return err
	}

	defer func() {
		if err := peer.SetDeadline(time.Time{}); err != nil { // clear write deadline
			log.WithError(err).Error("unable to clear deadline")
		}
	}()

	request := NewHello(self.identity.Token, self.headers)
	for k, v := range headers {
		request.Headers[k] = v
	}
	request.SetSequence(HelloSequence)
	if err := underlay.Tx(request); err != nil {
		_ = underlay.Close()
		return err
	}

	response, err := underlay.Rx()
	if err != nil {
		if errors.Is(err, BadMagicNumberError) {
			return fmt.Errorf("could not negotiate connection with %v, invalid header", peer.RemoteAddr().String())
		}
		return err
	}
	if !response.IsReplyingTo(HelloSequence) || response.ContentType != ContentTypeResultType {
		return fmt.Errorf("channel synchronization error, expected %v, got %v", HelloSequence, response.ReplyFor())
	}
	result := UnmarshalResult(response)
	if !result.Success {
		return errors.New(result.Message)
	}

	connectionId := string(headers[ConnectionIdHeader])
	if connectionId == "" {
		connectionId = string(response.Headers[ConnectionIdHeader])
	}
	id := ""

	if val, ok := response.GetStringHeader(IdHeader); ok {
		id = val
	} else if certs := underlay.Certificates(); len(certs) > 0 {
		id = certs[0].Subject.CommonName
	}

	// type should always be controller by the dialing side, the listener shouldn't be setting a type
	// in the header. Set the type here, so we can know the type on the dialing side as well
	response.Headers[TypeHeader] = headers[TypeHeader]
	underlay.init(id, connectionId, response.Headers)

	return nil
}
