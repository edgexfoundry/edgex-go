//go:build no_openziti

/*******************************************************************************
 * Copyright 2024 IOTech Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package zerotrust

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	btConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// HttpTransportFromService is the implementation for the case where the service is built with no_openziti flag.
func HttpTransportFromService(secretProvider interfaces.SecretProviderExt, _ config.ServiceInfo, lc logger.LoggingClient) (http.RoundTripper, error) {
	return httpDefaultTransport(secretProvider, lc)
}

// HttpTransportFromClient is the implementation for the case where the service is built with no_openziti flag.
func HttpTransportFromClient(secretProvider interfaces.SecretProviderExt, _ *config.ClientInfo, lc logger.LoggingClient) (http.RoundTripper, error) {
	return httpDefaultTransport(secretProvider, lc)
}

func httpDefaultTransport(secretProvider interfaces.SecretProviderExt, lc logger.LoggingClient) (http.RoundTripper, error) {
	if secretProvider.IsZeroTrustEnabled() {
		lc.Info("zero trust was enabled, but service is built with no_openziti flag. falling back to default HTTP transport")
	}
	return http.DefaultTransport, nil
}

// SetupWebListener is the implementation for the case where the service is built with no_openziti flag.
func SetupWebListener(serviceConfig config.ServiceInfo, serviceName, addr string, dic *di.Container) (net.Listener, error) {
	lc := container.LoggingClientFrom(dic.Get)
	listenMode, ok := serviceConfig.SecurityOptions[btConfig.SecurityModeKey]
	if ok {
		lc.Debugf("service security option %s = %s", btConfig.SecurityModeKey, listenMode)
		if strings.EqualFold(listenMode, ZeroTrustMode) {
			lc.Warnf("service %s is configured with zero trust security mode, but the service is built with no_openziti flag. all zero trust operations will be ignored.", serviceName)
		}
	}
	lc.Debugf("listening on underlay network. ListenMode '%s' at %s", listenMode, addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not listen on %s: %w", addr, err)
	}
	return listener, nil
}

// ListenOnMode is the implementation for the case where the service is built with no_openziti flag.
func ListenOnMode(bootstrapConfig config.BootstrapConfiguration, serverKey, addr string, _ startup.Timer, server *http.Server, dic *di.Container) error {
	lc := container.LoggingClientFrom(dic.Get)

	listenMode, ok := bootstrapConfig.Service.SecurityOptions[btConfig.SecurityModeKey]
	if ok && strings.EqualFold(listenMode, ZeroTrustMode) {
		lc.Warnf("service %s is configured with zero trust security mode, but the service is built with no_openziti flag. all zero trust operations will be ignored.", serverKey)
	}

	// following codes are executed when SecurityModeKey is not set or not equal to ZeroTrustMode
	lc.Infof("listening on underlay network. ListenMode '%s' at %s", listenMode, addr)
	ln, listenErr := net.Listen("tcp", addr)
	if listenErr != nil {
		return listenErr
	}
	return server.Serve(ln)
}
