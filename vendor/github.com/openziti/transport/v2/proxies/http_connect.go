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

package proxies

import (
	"bufio"
	"context"
	"github.com/michaelquigley/pfxlog"
	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

func NewHttpConnectProxyDialer(dialer proxy.Dialer, addr string, auth *proxy.Auth, timeout time.Duration) *HttpConnectProxyDialer {
	return &HttpConnectProxyDialer{
		dialer:  dialer,
		address: addr,
		auth:    auth,
		timeout: timeout,
	}
}

type HttpConnectProxyDialer struct {
	dialer  proxy.Dialer
	address string
	auth    *proxy.Auth
	timeout time.Duration
}

func (self *HttpConnectProxyDialer) Dial(network, addr string) (net.Conn, error) {
	dialer := self.dialer
	if dialer == nil {
		dialer = &net.Dialer{Timeout: self.timeout}
	}
	c, err := dialer.Dial(network, self.address)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to proxy server at %s", self.address)
	}

	if err = self.Connect(c, addr); err != nil {
		if closeErr := c.Close(); closeErr != nil {
			pfxlog.Logger().WithError(closeErr).Error("failed to close connection to proxy after connect error")
		}
		return nil, err
	}

	return c, nil
}

func (self *HttpConnectProxyDialer) Connect(c net.Conn, addr string) error {
	log := pfxlog.Logger()

	log.Debugf("create connect request to %s", addr)

	ctx := context.Background()
	if self.timeout > 0 {
		timeoutCtx, cancelF := context.WithTimeout(ctx, self.timeout)
		defer cancelF()
		ctx = timeoutCtx
	}

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: addr},
		Host:   addr,
		Header: http.Header{},
		Close:  false,
	}
	req = req.WithContext(ctx)
	if self.auth != nil {
		req.SetBasicAuth(self.auth.User, self.auth.Password)
	}
	req.Header.Set("User-Agent", "ziti-transport")

	if err := req.Write(c); err != nil {
		return errors.Wrapf(err, "unable to send connect request to proxy server at %s", self.address)
	}

	resp, err := http.ReadResponse(bufio.NewReader(c), req)
	if err != nil {
		return errors.Wrapf(err, "unable to read response to connect request to proxy server at %s", self.address)
	}

	defer func() {
		log.Debug("closing resp body")
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Errorf("proxy returned: %s", string(respBody))
		return errors.Errorf("received %v instead of 200 OK in response to connect request to proxy server at %s", resp.StatusCode, self.address)
	}

	return nil
}
