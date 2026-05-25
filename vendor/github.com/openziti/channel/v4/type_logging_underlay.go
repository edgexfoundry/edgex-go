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
	"github.com/michaelquigley/pfxlog"
	"net"
	"time"
)

type TypeLoggingUnderlay struct {
	wrapped Underlay
}

func (self *TypeLoggingUnderlay) Rx() (*Message, error) {
	return self.wrapped.Rx()
}

func (self *TypeLoggingUnderlay) Tx(m *Message) error {
	pfxlog.Logger().Infof("sending msg of type %d on %s\n", m.ContentType, GetUnderlayType(self.wrapped))
	return self.wrapped.Tx(m)
}

func (self *TypeLoggingUnderlay) Id() string {
	return self.wrapped.Id()
}

func (self *TypeLoggingUnderlay) LogicalName() string {
	return self.wrapped.LogicalName()
}

func (self *TypeLoggingUnderlay) ConnectionId() string {
	return self.wrapped.ConnectionId()
}

func (self *TypeLoggingUnderlay) Certificates() []*x509.Certificate {
	return self.wrapped.Certificates()
}

func (self *TypeLoggingUnderlay) Label() string {
	return self.wrapped.Label()
}

func (self *TypeLoggingUnderlay) Close() error {
	return self.wrapped.Close()
}

func (self *TypeLoggingUnderlay) IsClosed() bool {
	return self.wrapped.IsClosed()
}

func (self *TypeLoggingUnderlay) Headers() map[int32][]byte {
	return self.wrapped.Headers()
}

func (self *TypeLoggingUnderlay) SetWriteTimeout(duration time.Duration) error {
	return self.wrapped.SetWriteTimeout(duration)
}

func (self *TypeLoggingUnderlay) SetWriteDeadline(time time.Time) error {
	return self.wrapped.SetWriteDeadline(time)
}

func (self *TypeLoggingUnderlay) GetLocalAddr() net.Addr {
	return self.wrapped.GetLocalAddr()
}

func (self *TypeLoggingUnderlay) GetRemoteAddr() net.Addr {
	return self.wrapped.GetRemoteAddr()
}
