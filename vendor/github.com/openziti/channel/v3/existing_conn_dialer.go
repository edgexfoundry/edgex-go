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
	"net"
	"time"
)

type existingConnDialer struct {
	id      *identity.TokenId
	peer    net.Conn
	headers map[int32][]byte
}

func NewExistingConnDialer(id *identity.TokenId, peer net.Conn, headers map[int32][]byte) UnderlayFactory {
	return &existingConnDialer{
		id:      id,
		peer:    peer,
		headers: headers,
	}
}

func (self *existingConnDialer) Create(timeout time.Duration) (Underlay, error) {
	log := pfxlog.Logger()
	log.Debug("started")
	defer log.Debug("exited")

	version := uint32(2)
	tryCount := 0

	defer func() {
		if err := self.peer.SetDeadline(time.Time{}); err != nil { // clear write deadline
			log.WithError(err).Error("unable to clear write deadline")
		}
	}()

	for {
		impl := newExistingImpl(self.peer, version)

		if timeout > 0 {
			if err := self.peer.SetDeadline(time.Now().Add(timeout)); err != nil {
				return nil, err
			}
		}
		if err := self.sendHello(impl); err != nil {
			if tryCount > 0 {
				return nil, err
			} else {
				log.WithError(err).Warnf("error initiating channel with hello")
			}
			tryCount++
			if retryVersion, _ := GetRetryVersion(err); retryVersion != version {
				version = retryVersion
			} else {
				return nil, err
			}

			log.Warnf("Retrying dial with protocol version %v", version)
			continue
		}
		impl.id = self.id
		return impl, nil
	}
}

func (self *existingConnDialer) sendHello(impl *existingConnImpl) error {
	log := pfxlog.ContextLogger(impl.Label())
	defer log.Debug("exited")
	log.Debug("started")

	request := NewHello(self.id.Token, self.headers)
	request.sequence = HelloSequence
	if err := impl.Tx(request); err != nil {
		_ = impl.peer.Close()
		return err
	}

	response, err := impl.Rx()
	if err != nil {
		return err
	}
	if !response.IsReplyingTo(request.sequence) || response.ContentType != ContentTypeResultType {
		return fmt.Errorf("channel synchronization error, expected %v, got %v", request.sequence, response.ReplyFor())
	}
	result := UnmarshalResult(response)
	if !result.Success {
		return errors.New(result.Message)
	}
	impl.connectionId = string(response.Headers[ConnectionIdHeader])
	impl.headers = response.Headers

	return nil
}
