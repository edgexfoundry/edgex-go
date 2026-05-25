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
	"github.com/openziti/identity"
	"github.com/pkg/errors"
	"net"
	"time"
)

type existingConnListener struct {
	identity *identity.TokenId
	peer     net.Conn
	headers  map[int32][]byte
}

func NewExistingConnListener(identity *identity.TokenId, peer net.Conn, headers map[int32][]byte) UnderlayFactory {
	return &existingConnListener{
		identity: identity,
		peer:     peer,
		headers:  headers,
	}
}

func (self *existingConnListener) Create(timeout time.Duration) (Underlay, error) {
	log := pfxlog.Logger()

	impl := newExistingImpl(self.peer, 2)
	connectionId, err := NextConnectionId()
	if err != nil {
		return nil, errors.Wrap(err, "error getting connection id")
	}
	impl.connectionId = connectionId

	if timeout > 0 {
		defer func() {
			if err = self.peer.SetDeadline(time.Time{}); err != nil {
				log.WithError(err).Error("unable to clear deadline on conn after create")
			}
		}()

		if err = self.peer.SetDeadline(time.Now().Add(timeout)); err != nil {
			return nil, errors.Wrap(err, "could not set deadline on conn")
		}
	}

	request, hello, err := self.receiveHello(impl)
	if err != nil {
		return nil, errors.Wrap(err, "error receiving hello")
	}

	impl.id = &identity.TokenId{Token: hello.IdToken}
	impl.headers = hello.Headers

	if err = self.ackHello(impl, request, true, ""); err != nil {
		return nil, errors.Wrap(err, "unable to acknowledge hello")
	}

	return impl, nil
}

func (self *existingConnListener) receiveHello(impl *existingConnImpl) (*Message, *Hello, error) {
	log := pfxlog.ContextLogger(impl.Label())
	log.Debug("started")
	defer log.Debug("exited")

	request, err := impl.rxHello()
	if err != nil {
		if err == BadMagicNumberError {
			WriteUnknownVersionResponse(impl.peer)
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

func (self *existingConnListener) ackHello(impl *existingConnImpl, request *Message, success bool, message string) error {
	response := NewResult(success, message)

	for key, val := range self.headers {
		response.Headers[key] = val
	}

	response.Headers[ConnectionIdHeader] = []byte(impl.connectionId)
	response.sequence = HelloSequence

	response.ReplyTo(request)
	return impl.Tx(response)
}
