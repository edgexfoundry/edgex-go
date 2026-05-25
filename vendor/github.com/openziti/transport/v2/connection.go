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

package transport

import (
	"crypto/x509"
	"net"
)

type ConnectionDetail struct {
	Address string
	InBound bool
	Name    string
}

func (cd *ConnectionDetail) String() string {
	out := ""
	if cd.InBound {
		out += cd.Address + " <-"
		if len(cd.Name) > 0 {
			out += " {" + cd.Name + "}"
		}

	} else {
		if len(cd.Name) > 0 {
			out += "{" + cd.Name + "} "
		}
		out += "-> " + cd.Address
	}
	return out
}

// Conn represents an abstract connection
type Conn interface {
	Detail() *ConnectionDetail
	PeerCertificates() []*x509.Certificate
	net.Conn
}
