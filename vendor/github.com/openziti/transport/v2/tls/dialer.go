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

package tls

import (
	"crypto/tls"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
	"github.com/openziti/transport/v2/proxies"
	"github.com/pkg/errors"
	"time"
)

func Dial(a address, name string, i *identity.TokenId, timeout time.Duration, proxyConf *transport.ProxyConfiguration, protocols ...string) (transport.Conn, error) {
	return DialWithLocalBinding(a, name, "", i, timeout, proxyConf, protocols...)
}

func DialWithLocalBinding(a address, name, localBinding string, i *identity.TokenId, timeout time.Duration, proxyConf *transport.ProxyConfiguration, protocols ...string) (transport.Conn, error) {
	destination := a.bindableAddress()
	dialer, err := transport.NewDialerWithLocalBinding("tcp", timeout, localBinding)
	if err != nil {
		return nil, err
	}

	log := pfxlog.Logger().WithField("dest", destination)

	tlsCfg := i.ClientTLSConfig()
	tlsCfg.ServerName = a.hostname
	if len(protocols) > 0 {
		tlsCfg = tlsCfg.Clone()
		tlsCfg.NextProtos = append(tlsCfg.NextProtos, protocols...)
	}

	var tlsConn *tls.Conn

	if proxyConf != nil && proxyConf.Type != transport.ProxyTypeNone {
		if proxyConf.Type == transport.ProxyTypeHttpConnect {
			log.Infof("using http connect proxy at %s", proxyConf.Address)
			proxyDialer := proxies.NewHttpConnectProxyDialer(dialer, proxyConf.Address, proxyConf.Auth, timeout)
			conn, err := proxyDialer.Dial("tcp", destination)
			if err != nil {
				return nil, err
			}

			tlsConn = tls.Client(conn, tlsCfg)
		} else {
			return nil, errors.Errorf("unsupported proxy type %s", string(proxyConf.Type))
		}
	} else {
		tlsConn, err = tls.DialWithDialer(dialer, "tcp", destination, tlsCfg)
		if err != nil {
			return nil, err
		}
	}

	log.Debugf("server provided [%d] certificates", len(tlsConn.ConnectionState().PeerCertificates))

	return &Connection{
		detail: &transport.ConnectionDetail{
			Address: Type + ":" + destination,
			InBound: false,
			Name:    name,
		},
		Conn: tlsConn,
	}, nil
}
