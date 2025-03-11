/*
	Copyright 2019 NetFoundry Inc.

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

package edge_apis

import (
	"crypto/x509"
	"github.com/go-openapi/runtime"
	openapiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_client_api_client"
	"github.com/openziti/edge-api/rest_management_api_client"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

// ApiType is an interface constraint for generics. The underlying go-swagger types only have fields, which are
// insufficient to attempt to make a generic type from. Instead, this constraint is used that points at the
// aliased types.
type ApiType interface {
	ZitiEdgeManagement | ZitiEdgeClient
}

type OidcEnabledApi interface {
	// SetUseOidc forces an API Client to operate in OIDC mode (true) or legacy mode (false). The state of the controller
	// is ignored and dynamic enable/disable of OIDC support is suspended.
	SetUseOidc(use bool)

	// SetAllowOidcDynamicallyEnabled sets whether clients will check the controller for OIDC support or not. If supported
	// OIDC is favored over legacy authentication.
	SetAllowOidcDynamicallyEnabled(allow bool)
}

// BaseClient implements the Client interface specifically for the types specified in the ApiType constraint. It
// provides shared functionality that all ApiType types require.
type BaseClient[A ApiType] struct {
	API            *A
	AuthEnabledApi AuthEnabledApi
	Components
	AuthInfoWriter        runtime.ClientAuthInfoWriter
	ApiSession            atomic.Pointer[ApiSession]
	Credentials           Credentials
	ApiUrls               []*url.URL
	ApiBinding            string
	ApiVersion            string
	Schemes               []string
	onControllerListeners []func([]*url.URL)
}

func (self *BaseClient[A]) Url() url.URL {
	return *self.AuthEnabledApi.GetClientTransportPool().GetActiveTransport().ApiUrl
}

func (self *BaseClient[A]) AddOnControllerUpdateListeners(listener func([]*url.URL)) {
	self.onControllerListeners = append(self.onControllerListeners, listener)
}

// GetCurrentApiSession returns the ApiSession that is being used to authenticate requests.
func (self *BaseClient[A]) GetCurrentApiSession() ApiSession {
	ptr := self.ApiSession.Load()
	if ptr == nil {
		return nil
	}

	return *ptr
}

func (self *BaseClient[A]) SetUseOidc(use bool) {
	v := any(self.API)
	apiType := v.(OidcEnabledApi)
	apiType.SetUseOidc(use)
}

func (self *BaseClient[A]) SetAllowOidcDynamicallyEnabled(allow bool) {
	v := any(self.API)
	apiType := v.(OidcEnabledApi)
	apiType.SetAllowOidcDynamicallyEnabled(allow)
}

// Authenticate will attempt to use the provided credentials to authenticate via the underlying ApiType. On success
// the API Session details will be returned and the current client will make authenticated requests on future
// calls. On an error the API Session in use will be cleared and subsequent requests will become/continue to be
// made in an unauthenticated fashion.
func (self *BaseClient[A]) Authenticate(credentials Credentials, configTypesOverride []string) (ApiSession, error) {

	self.Credentials = nil
	self.ApiSession.Store(nil)

	if credCaPool := credentials.GetCaPool(); credCaPool != nil {
		self.HttpTransport.TLSClientConfig.RootCAs = credCaPool
	} else {
		self.HttpTransport.TLSClientConfig.RootCAs = self.Components.CaPool
	}

	apiSession, err := self.AuthEnabledApi.Authenticate(credentials, configTypesOverride, self.HttpClient)

	if err != nil {
		return nil, err
	}

	self.Credentials = credentials
	self.ApiSession.Store(&apiSession)

	self.ProcessControllers(self.AuthEnabledApi)

	return apiSession, nil
}

// initializeComponents assembles the lower level components necessary for the go-swagger/openapi facilities.
func (self *BaseClient[A]) initializeComponents(config *ApiClientConfig) {
	components := NewComponentsWithConfig(&ComponentsConfig{
		Proxy: config.Proxy,
	})
	components.HttpTransport.TLSClientConfig.RootCAs = config.CaPool
	components.CaPool = config.CaPool

	self.Components = *components
}

func NewRuntime(apiUrl *url.URL, schemes []string, httpClient *http.Client) *openapiclient.Runtime {
	return openapiclient.NewWithClient(apiUrl.Host, apiUrl.Path, schemes, httpClient)
}

// AuthenticateRequest implements the openapi runtime.ClientAuthInfoWriter interface from the OpenAPI libraries. It is used
// to authenticate outgoing requests.
func (self *BaseClient[A]) AuthenticateRequest(request runtime.ClientRequest, registry strfmt.Registry) error {
	if self.AuthInfoWriter != nil {
		return self.AuthInfoWriter.AuthenticateRequest(request, registry)
	}

	// do not add auth to authenticating endpoints
	if strings.Contains(request.GetPath(), "/oidc/auth") || strings.Contains(request.GetPath(), "/authenticate") {
		return nil
	}

	currentSessionPtr := self.ApiSession.Load()
	if currentSessionPtr != nil {
		currentSession := *currentSessionPtr

		if currentSession != nil && currentSession.GetToken() != nil {
			if err := currentSession.AuthenticateRequest(request, registry); err != nil {
				return err
			}
		}
	}

	if self.Credentials != nil {
		if err := self.Credentials.AuthenticateRequest(request, registry); err != nil {
			return err
		}
	}

	return nil
}

func (self *BaseClient[A]) ProcessControllers(authEnabledApi AuthEnabledApi) {
	list, err := authEnabledApi.ListControllers()

	if err != nil {
		pfxlog.Logger().WithError(err).Error("error listing controllers, continuing with 1 default configured controller")
		return
	}

	if list == nil || len(*list) <= 1 {
		pfxlog.Logger().Debug("no additional controllers reported, continuing with 1 default configured controller")
		return
	}

	//look for matching api binding and versions
	for _, controller := range *list {
		apis := controller.APIAddresses[self.ApiBinding]

		for _, apiAddr := range apis {
			if apiAddr.Version == self.ApiVersion {
				apiUrl, parseErr := url.Parse(apiAddr.URL)
				if parseErr == nil {
					self.AuthEnabledApi.GetClientTransportPool().Add(apiUrl, NewRuntime(apiUrl, self.Schemes, self.HttpClient))
				}
			}
		}
	}

	apis := self.AuthEnabledApi.GetClientTransportPool().GetApiUrls()
	for _, listener := range self.onControllerListeners {
		listener(apis)
	}
}

// ManagementApiClient provides the ability to authenticate and interact with the Edge Management API.
type ManagementApiClient struct {
	BaseClient[ZitiEdgeManagement]
}

type ApiClientConfig struct {
	ApiUrls      []*url.URL
	CaPool       *x509.CertPool
	TotpCallback func(chan string)
	Proxy        func(r *http.Request) (*url.URL, error)
}

// NewManagementApiClient will assemble an ManagementApiClient. The apiUrl should be the full URL
// to the Edge Management API (e.g. `https://example.com/edge/management/v1`).
//
// The `caPool` argument should be a list of trusted root CAs. If provided as `nil` here unauthenticated requests
// will use the system certificate pool. If authentication occurs, and a certificate pool is set on the Credentials
// the certificate pool from the Credentials will be used from that point forward. Credentials implementations
// based on an identity.Identity are likely to provide a certificate pool.
//
// For OpenZiti instances not using publicly signed certificates, `ziti.GetControllerWellKnownCaPool()` can be used
// to obtain and verify the target controllers CAs. Tools should allow users to verify and accept new controllers
// that have not been verified from an outside secret (such as an enrollment token).
func NewManagementApiClient(apiUrls []*url.URL, caPool *x509.CertPool, totpCallback func(chan string)) *ManagementApiClient {
	return NewManagementApiClientWithConfig(&ApiClientConfig{
		ApiUrls:      apiUrls,
		CaPool:       caPool,
		TotpCallback: totpCallback,
		Proxy:        http.ProxyFromEnvironment,
	})
}

func NewManagementApiClientWithConfig(config *ApiClientConfig) *ManagementApiClient {
	ret := &ManagementApiClient{}
	ret.Schemes = rest_management_api_client.DefaultSchemes
	ret.ApiBinding = "edge-management"
	ret.ApiVersion = "v1"
	ret.ApiUrls = config.ApiUrls
	ret.initializeComponents(config)

	transportPool := NewClientTransportPoolRandom()

	for _, apiUrl := range config.ApiUrls {
		newRuntime := NewRuntime(apiUrl, ret.Schemes, ret.Components.HttpClient)
		newRuntime.DefaultAuthentication = ret
		transportPool.Add(apiUrl, newRuntime)
	}

	newApi := rest_management_api_client.New(transportPool, nil)
	api := ZitiEdgeManagement{
		ZitiEdgeManagement:  newApi,
		TotpCallback:        config.TotpCallback,
		ClientTransportPool: transportPool,
	}

	ret.API = &api
	ret.AuthEnabledApi = &api

	return ret
}

type ClientApiClient struct {
	BaseClient[ZitiEdgeClient]
}

// NewClientApiClient will assemble a  ClientApiClient. The apiUrl should be the full URL
// to the Edge Client API (e.g. `https://example.com/edge/client/v1`).
//
// The `caPool` argument should be a list of trusted root CAs. If provided as `nil` here unauthenticated requests
// will use the system certificate pool. If authentication occurs, and a certificate pool is set on the Credentials
// the certificate pool from the Credentials will be used from that point forward. Credentials implementations
// based on an identity.Identity are likely to provide a certificate pool.
//
// For OpenZiti instances not using publicly signed certificates, `ziti.GetControllerWellKnownCaPool()` can be used
// to obtain and verify the target controllers CAs. Tools should allow users to verify and accept new controllers
// that have not been verified from an outside secret (such as an enrollment token).
func NewClientApiClient(apiUrls []*url.URL, caPool *x509.CertPool, totpCallback func(chan string)) *ClientApiClient {
	return NewClientApiClientWithConfig(&ApiClientConfig{
		ApiUrls:      apiUrls,
		CaPool:       caPool,
		TotpCallback: totpCallback,
		Proxy:        http.ProxyFromEnvironment,
	})
}

func NewClientApiClientWithConfig(config *ApiClientConfig) *ClientApiClient {
	ret := &ClientApiClient{}
	ret.ApiBinding = "edge-client"
	ret.ApiVersion = "v1"
	ret.Schemes = rest_client_api_client.DefaultSchemes
	ret.ApiUrls = config.ApiUrls

	ret.initializeComponents(config)

	transportPool := NewClientTransportPoolRandom()

	for _, apiUrl := range config.ApiUrls {
		newRuntime := NewRuntime(apiUrl, ret.Schemes, ret.Components.HttpClient)
		newRuntime.DefaultAuthentication = ret
		transportPool.Add(apiUrl, newRuntime)
	}

	newApi := rest_client_api_client.New(transportPool, nil)
	api := ZitiEdgeClient{
		ZitiEdgeClient:      newApi,
		TotpCallback:        config.TotpCallback,
		ClientTransportPool: transportPool,
	}
	ret.API = &api
	ret.AuthEnabledApi = &api

	return ret
}
