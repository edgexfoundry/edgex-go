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
	"fmt"
	"github.com/pkg/errors"
	"net"
	"time"
)

// NewDialerWithLocalBinding creates a dialer and sets the local ip used for dialing
func NewDialerWithLocalBinding(addressType string, timeout time.Duration, localBinding string) (*net.Dialer, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	ip, err := ResolveLocalBinding(localBinding)
	if err != nil {
		return nil, err
	}

	if ip != nil {
		switch addressType {
		case "udp":
			dialer.LocalAddr = &net.UDPAddr{
				IP: ip,
			}
		case "tcp", "tls":
			dialer.LocalAddr = &net.TCPAddr{
				IP: ip,
			}
		default:
			return nil, errors.New(fmt.Sprintf("Unsupported addressType: %s", addressType))
		}
	}

	return dialer, nil
}

func ResolveLocalBinding(localBinding string) (net.IP, error) {
	if localBinding != "" {
		iface, err := ResolveInterface(localBinding)

		if err != nil {
			return nil, err
		}

		addrs, err := iface.Addrs()

		if err != nil {
			return nil, err
		}

		if len(addrs) == 0 {
			return nil, errors.New(fmt.Sprintf("no ip addresses assigned to interface %s", localBinding))
		}

		return addrs[0].(*net.IPNet).IP, nil
	}

	return nil, nil
}
