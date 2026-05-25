package edge_apis

import (
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/go-openapi/runtime"
	openapiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_model"
)

const (
	AuthRequestIdHeader = "auth-request-id"
	TotpRequiredHeader  = "totp-required"
)

// AuthEnabledApi is a sentinel interface that detects APIs supporting authentication.
// It provides methods for authenticating, managing sessions, and discovering controllers for high-availability.
type AuthEnabledApi interface {
	// Authenticate authenticates using the provided credentials and returns an ApiSession for subsequent authenticated requests.
	Authenticate(credentials Credentials, configTypes []string, httpClient *http.Client) (ApiSession, error)
	// SetUseOidc forces OIDC mode (true) or legacy mode (false).
	SetUseOidc(bool)
	// ListControllers returns the list of available controllers for HA failover.
	ListControllers() (*rest_model.ControllersList, error)
	// GetClientTransportPool returns the transport pool managing multiple controller endpoints.
	GetClientTransportPool() ClientTransportPool
	// SetClientTransportPool sets the transport pool.
	SetClientTransportPool(ClientTransportPool)
	// RefreshApiSession refreshes an existing session.
	RefreshApiSession(apiSession ApiSession, httpClient *http.Client) (ApiSession, error)
}

// BaseClient provides shared authentication and session management for OpenZiti API clients.
// It handles credential-based authentication, TLS configuration, session storage, and controller failover.
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

// Url returns the URL of the currently active controller endpoint.
func (self *BaseClient[A]) Url() url.URL {
	return *self.AuthEnabledApi.GetClientTransportPool().GetActiveTransport().ApiUrl
}

// AddOnControllerUpdateListeners registers a callback that is invoked when the list of
// available controller endpoints changes.
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

// SetUseOidc forces the API client to operate in OIDC mode when true, or legacy mode when false.
func (self *BaseClient[A]) SetUseOidc(use bool) {
	v := any(self.API)
	apiType := v.(OidcEnabledApi)
	apiType.SetUseOidc(use)
}

// SetAllowOidcDynamicallyEnabled configures whether the client checks the controller for
// OIDC support and switches modes accordingly.
func (self *BaseClient[A]) SetAllowOidcDynamicallyEnabled(allow bool) {
	v := any(self.API)
	apiType := v.(OidcEnabledApi)
	apiType.SetAllowOidcDynamicallyEnabled(allow)
}

func (self *BaseClient[A]) SetOidcRedirectUri(redirectUri string) {
	v := any(self.API)
	apiType := v.(OidcEnabledApi)
	apiType.SetOidcRedirectUri(redirectUri)
}

// Authenticate authenticates using provided credentials, updating the TLS configuration based on the credential's CA pool.
// On success, stores the session and processes controller endpoints for HA failover.
// On failure, clears the session and credentials.
func (self *BaseClient[A]) Authenticate(credentials Credentials, configTypesOverride []string) (ApiSession, error) {
	self.Credentials = nil
	self.ApiSession.Store(nil)

	tlsClientConfig := self.TlsAwareTransport.GetTlsClientConfig()

	if credCaPool := credentials.GetCaPool(); credCaPool != nil {
		tlsClientConfig.RootCAs = credCaPool
	} else {
		tlsClientConfig.RootCAs = self.CaPool
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

func (self *BaseClient[A]) AuthenticateWithPreviousSession(credentials Credentials, prevApiSession ApiSession) (ApiSession, error) {
	self.Credentials = nil
	self.ApiSession.Store(nil)

	tlsClientConfig := self.TlsAwareTransport.GetTlsClientConfig()
	if credCaPool := credentials.GetCaPool(); credCaPool != nil {
		tlsClientConfig.RootCAs = credCaPool
	} else {
		tlsClientConfig.RootCAs = self.CaPool
	}

	refreshedSession, refreshErr := self.AuthEnabledApi.RefreshApiSession(prevApiSession, self.HttpClient)

	if refreshErr != nil {
		return nil, refreshErr
	}

	self.Credentials = credentials
	self.ApiSession.Store(&refreshedSession)

	self.ProcessControllers(self.AuthEnabledApi)

	return refreshedSession, nil
}

// initializeComponents assembles HTTP client infrastructure, either using provided Components or creating new ones.
// If Components are provided with nil transport/client, they are initialized with warnings logged.
func (self *BaseClient[A]) initializeComponents(config *ApiClientConfig) {
	//have a config and either the client or transport are set, verify them, else an empty components was supplied
	// then initialize them with defaults
	if config.Components != nil && (config.Components.HttpClient != nil || config.Components.TlsAwareTransport != nil) {
		config.Components.assertComponents(config)
		self.Components = *config.Components

		if config.Proxy != nil {
			pfxlog.Logger().Warn("components were provided along with a proxy function on the ApiClientConfig, it is being ignored, if needed properly set on components")
		}
		return
	}

	components := NewComponentsWithConfig(&ComponentsConfig{
		Proxy: config.Proxy,
	})

	tlsClientConfig := components.TlsAwareTransport.GetTlsClientConfig()
	tlsClientConfig.RootCAs = config.CaPool
	components.CaPool = config.CaPool

	self.Components = *components
}

// NewRuntime creates an OpenAPI runtime for communicating with a controller endpoint. Used for HA failover to add multiple controller endpoints.
func NewRuntime(apiUrl *url.URL, schemes []string, httpClient *http.Client) *openapiclient.Runtime {
	return openapiclient.NewWithClient(apiUrl.Host, apiUrl.Path, schemes, httpClient)
}

// AuthenticateRequest authenticates outgoing API requests using the current session or credentials.
// It implements the openapi runtime.ClientAuthInfoWriter interface.
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

// ProcessControllers discovers peer controllers and registers them for HA failover. Called after successful authentication.
func (self *BaseClient[A]) ProcessControllers(authEnabledApi AuthEnabledApi) {
	list, err := authEnabledApi.ListControllers()

	if err != nil {
		pfxlog.Logger().WithError(err).Debug("error listing controllers, continuing with 1 default configured controller")
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
