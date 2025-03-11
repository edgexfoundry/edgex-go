//go:build !no_openziti

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
	"context"
	"errors"
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
	edgeapis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
)

func authToOpenZiti(ozController, jwt string) (ziti.Context, error) {
	if !strings.Contains(ozController, "://") {
		ozController = "https://" + ozController
	}
	caPool, caErr := ziti.GetControllerWellKnownCaPool(ozController)
	if caErr != nil {
		return nil, caErr
	}

	credentials := edgeapis.NewJwtCredentials(jwt)
	credentials.CaPool = caPool

	cfg := &ziti.Config{
		ZtAPI:       ozController + "/edge/client/v1",
		Credentials: credentials,
	}
	cfg.ConfigTypes = append(cfg.ConfigTypes, "all")

	ctx, ctxErr := ziti.NewContext(cfg)
	if ctxErr != nil {
		return nil, ctxErr
	}
	if authErr := ctx.Authenticate(); authErr != nil {
		return nil, authErr
	}

	return ctx, nil
}

func HttpTransportFromService(secretProvider interfaces.SecretProviderExt, serviceInfo config.ServiceInfo, lc logger.LoggingClient) (http.RoundTripper, error) {
	roundTripper := http.DefaultTransport
	if secretProvider.IsZeroTrustEnabled() {
		lc.Debugf("zero trust client detected for service: %s", serviceInfo.Host)
		if rt, err := createZitifiedTransport(secretProvider, serviceInfo.SecurityOptions[OpenZitiControllerKey]); err != nil {
			return nil, err
		} else {
			roundTripper = rt
		}
	}
	return roundTripper, nil
}

func HttpTransportFromClient(secretProvider interfaces.SecretProviderExt, clientInfo *config.ClientInfo, lc logger.LoggingClient) (http.RoundTripper, error) {
	roundTripper := http.DefaultTransport
	if secretProvider.IsZeroTrustEnabled() {
		lc.Debugf("zero trust client detected for client: %s", clientInfo.Host)
		if rt, err := createZitifiedTransport(secretProvider, clientInfo.SecurityOptions[OpenZitiControllerKey]); err != nil {
			return nil, err
		} else {
			roundTripper = rt
		}
	}
	return roundTripper, nil
}

type ZitiDialer struct {
	underlayDialer *net.Dialer
}

func (z ZitiDialer) Dial(network, address string) (net.Conn, error) {
	return z.underlayDialer.Dial(network, address)
}

func createZitifiedTransport(secretProvider interfaces.SecretProviderExt, ozController string) (http.RoundTripper, error) {
	jwt, errJwt := secretProvider.GetSelfJWT()
	if errJwt != nil {
		return nil, fmt.Errorf("could not load jwt: %v", errJwt)
	}
	ctx, authErr := authToOpenZiti(ozController, jwt)
	if authErr != nil {
		return nil, fmt.Errorf("could not authenticate to OpenZiti: %v", authErr)
	}

	zitiContexts := ziti.NewSdkCollection()
	zitiContexts.Add(ctx)

	fallback := &ZitiDialer{
		underlayDialer: secretProvider.FallbackDialer(),
	}
	zitiTransport := http.DefaultTransport.(*http.Transport).Clone() // copy default transport
	zitiTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := zitiContexts.NewDialerWithFallback(ctx, fallback)
		return dialer.Dial(network, addr)
	}
	return zitiTransport, nil
}

func validateOpenZitiOptions(serviceConfig config.ServiceInfo, dic *di.Container) (ozToken, ozUrl string, err error) {
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	if secretProvider == nil {
		return ozToken, ozUrl, errors.New("nil secret provider is nil. cannot setup web listener for zero trust overlay network")
	}
	secretProvider.EnableZeroTrust() //mark the secret provider as zero trust enabled
	ozToken, err = secretProvider.GetSelfJWT()
	if err != nil {
		return ozToken, ozUrl, fmt.Errorf("could not load jwt from secret provider while setting up web listener for zero trust overlay network: %w", err)
	}
	ozUrl, exists := serviceConfig.SecurityOptions[OpenZitiControllerKey]
	if !exists {
		return ozToken, ozUrl, fmt.Errorf("%s is not set in the service security options under zero trust mode", OpenZitiControllerKey)
	}
	return ozToken, ozUrl, nil
}

func SetupWebListener(serviceConfig config.ServiceInfo, serviceName, addr string, dic *di.Container) (net.Listener, error) {
	lc := container.LoggingClientFrom(dic.Get)
	listenMode, ok := serviceConfig.SecurityOptions[btConfig.SecurityModeKey]
	if ok {
		lc.Debugf("service security option %s = %s", btConfig.SecurityModeKey, listenMode)
		if strings.EqualFold(listenMode, ZeroTrustMode) {
			ozToken, ozUrl, err := validateOpenZitiOptions(serviceConfig, dic)
			if err != nil {
				return nil, fmt.Errorf("could not setup web listener for zero trust overlay network: %w", err)
			}
			zitiCtx, err := authToOpenZiti(ozUrl, ozToken)
			if err != nil {
				return nil, fmt.Errorf("could not authenticate to OpenZiti while preparing web listner: %w", err)
			}
			ozServiceName := OpenZitiServicePrefix + serviceName
			lc.Debugf("Using OpenZiti service name: %s", ozServiceName)
			lc.Debugf("listening on overlay network. ListenMode '%s' at %s", listenMode, addr)
			listener, err := zitiCtx.Listen(ozServiceName)
			if err != nil {
				return nil, fmt.Errorf("could not bind service %s: %w", ozServiceName, err)
			}
			return listener, nil
		}
	}
	lc.Debugf("listening on underlay network. ListenMode '%s' at %s", listenMode, addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not listen on %s: %w", addr, err)
	}
	return listener, nil
}

type OpenZitiIdentityKey struct{}

// ListenOnMode configures and starts an HTTP server based on the provided security mode.
// If the security mode is set to zerotrust, it establishes a zero trust overlay network using OpenZiti.
// Otherwise, it listens on the specified address using the default network.
func ListenOnMode(bootstrapConfig config.BootstrapConfiguration, serverKey, addr string, t startup.Timer, server *http.Server, dic *di.Container) error {
	lc := container.LoggingClientFrom(dic.Get)

	server.ConnContext = mutator

	listenMode, ok := bootstrapConfig.Service.SecurityOptions[btConfig.SecurityModeKey]
	if ok {
		lc.Debugf("service security option %s = %s", btConfig.SecurityModeKey, listenMode)
		if strings.EqualFold(listenMode, ZeroTrustMode) {
			ozToken, ozUrl, err := validateOpenZitiOptions(*bootstrapConfig.Service, dic)
			if err != nil {
				return fmt.Errorf("could not setup web listener for zero trust overlay network: %w", err)
			}
			if !strings.Contains(ozUrl, "://") {
				ozUrl = "https://" + ozUrl
			}
			caPool, caErr := ziti.GetControllerWellKnownCaPool(ozUrl)
			if caErr != nil {
				return fmt.Errorf("fail to get CA pool while establishing zero trust overlay network: %w", caErr)
			}

			credentials := edgeapis.NewJwtCredentials(ozToken)
			credentials.CaPool = caPool

			cfg := &ziti.Config{
				ZtAPI:       ozUrl + "/edge/client/v1",
				Credentials: credentials,
			}
			cfg.ConfigTypes = append(cfg.ConfigTypes, "all")

			zitiCtx, ctxErr := ziti.NewContext(cfg)
			if ctxErr != nil {
				return fmt.Errorf("fail to create OpenZiti context while establishing zero trust overlay network: %w", ctxErr)
			}

			ozServiceName := OpenZitiServicePrefix + serverKey
			lc.Infof("Using OpenZiti service name: %s", ozServiceName)
			for t.HasNotElapsed() {
				ln, listenErr := zitiCtx.Listen(ozServiceName)
				if listenErr != nil {
					lc.Errorf("fail to bind OpenZiti service %s: %s. wait for later retry", ozServiceName, listenErr.Error())
					t.SleepForInterval()
				} else {
					lc.Infof("listening on OpenZiti overlay network at %s", addr)
					return server.Serve(ln)
				}
			}
			return errors.New("could not listen on the OpenZiti overlay network. timeout reached")
		}
	}
	// following codes are executed when SecurityModeKey is not set or not equal to ZeroTrustMode
	lc.Infof("listening on underlay network. ListenMode '%s' at %s", listenMode, addr)
	ln, listenErr := net.Listen("tcp", addr)
	if listenErr != nil {
		return listenErr
	}
	return server.Serve(ln)
}

func mutator(srcCtx context.Context, c net.Conn) context.Context {
	if zitiConn, ok := c.(edge.Conn); ok {
		return context.WithValue(srcCtx, OpenZitiIdentityKey{}, zitiConn)
	}
	return srcCtx
}
