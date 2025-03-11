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
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
)

var _ transport.Address = &address{} // enforce that address implements transport.Address

const Type = "wss"

type address struct {
	hostname string
	port     uint16
}

func (a address) Dial(name string, i *identity.TokenId, t time.Duration, c transport.Configuration) (transport.Conn, error) {
	u := url.URL{Scheme: "wss", Host: a.bindableAddress(), Path: "/ws"}
	return Dial(name, u, i, t, c)
}

func (a address) DialWithLocalBinding(name string, localBinding string, i *identity.TokenId, t time.Duration, c transport.Configuration) (transport.Conn, error) {
	u := url.URL{Scheme: "wss", Host: a.bindableAddress(), Path: "/ws"}
	return DialWithLocalBinding(name, u, localBinding, i, t, c)
}

func (a address) Listen(name string, i *identity.TokenId, acceptF func(transport.Conn), tcfg transport.Configuration) (io.Closer, error) {
	var wssConfig map[interface{}]interface{}
	if tcfg != nil {
		if v, found := tcfg[Type]; found {
			wssConfig = v.(map[interface{}]interface{})
		}
	}
	return Listen(a.bindableAddress(), name, i, acceptF, wssConfig)
}

func (a address) MustListen(name string, i *identity.TokenId, acceptF func(transport.Conn), tcfg transport.Configuration) io.Closer {
	closer, err := a.Listen(name, i, acceptF, tcfg)
	if err != nil {
		panic(err)
	}
	return closer
}

func (a address) String() string {
	return fmt.Sprintf("%s:%s", Type, a.bindableAddress())
}

func (a address) bindableAddress() string {
	return fmt.Sprintf("%s:%d", a.hostname, a.port)
}

func (a address) Type() string {
	return Type
}

func (a address) Hostname() string {
	return a.hostname
}

func (a address) Port() uint16 {
	return a.port
}

type AddressParser struct{}

func (ap AddressParser) Parse(s string) (transport.Address, error) {
	tokens := strings.Split(s, ":")
	if len(tokens) < 2 {
		return nil, errors.New("invalid format")
	}

	if tokens[0] == Type {
		if len(tokens) != 3 {
			return nil, errors.New("invalid format")
		}

		port, err := strconv.ParseUint(tokens[2], 10, 16)
		if err != nil {
			return nil, err
		}

		return &address{hostname: tokens[1], port: uint16(port)}, nil
	}

	return nil, errors.New("invalid format")
}
