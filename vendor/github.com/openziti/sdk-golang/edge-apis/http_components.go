package edge_apis

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_util"
)

// Components provides the foundational HTTP client infrastructure for OpenAPI clients,
// bundling the HTTP client, transport, and certificate pool as a cohesive unit.
type Components struct {
	HttpClient        *http.Client
	TlsAwareTransport TlsAwareTransport
	CaPool            *x509.CertPool
}

// assertComponents ensures that the components are initialized properly.
func (c Components) assertComponents(config *ApiClientConfig) {
	if config.Components.HttpClient == nil {
		pfxlog.Logger().Warn("components were provided but the http client was nil, it is being initialized")

		if config.Components.TlsAwareTransport == nil {
			config.Components.TlsAwareTransport = NewTlsAwareHttpTransport(nil)
			pfxlog.Logger().Warn("components were provided but the client and transport are nil, they are being initialized with a default")
		}

		config.Components.HttpClient = NewHttpClient(config.Components.TlsAwareTransport)
	}

	if config.Components.TlsAwareTransport == nil {
		if tlsAwareTransport, ok := config.Components.HttpClient.Transport.(TlsAwareTransport); ok {
			config.Components.TlsAwareTransport = tlsAwareTransport
			pfxlog.Logger().Warn("components were provided but the transport was nil, it is being initialized with the transport from the http client")
		} else {
			pfxlog.Logger().Warn("components were provided but the transport was nil and the client did not have a suitable transport, it is being initialized with a default")
			config.Components.TlsAwareTransport = NewTlsAwareHttpTransport(nil)
			config.Components.HttpClient.Transport = config.Components.TlsAwareTransport
		}
	}

	if config.Components.HttpClient.Transport != config.Components.TlsAwareTransport {
		pfxlog.Logger().Warn("components were provided but the http client transport was not the same as the transport in components, it is being initialized")
		config.Components.HttpClient.Transport = config.Components.TlsAwareTransport
	}
}

// ComponentsConfig contains configuration options for creating Components.
type ComponentsConfig struct {
	Proxy func(*http.Request) (*url.URL, error)
}

// NewComponentsWithConfig assembles a new set of components using the provided configuration.
func NewComponentsWithConfig(cfg *ComponentsConfig) *Components {
	tlsAwareHttpTransport := NewTlsAwareHttpTransport(cfg)
	httpClient := NewHttpClient(tlsAwareHttpTransport)

	return &Components{
		HttpClient:        httpClient,
		TlsAwareTransport: tlsAwareHttpTransport,
	}
}

// NewHttpClient creates an HTTP client with the given transport.
func NewHttpClient(tlsAwareHttpTransport TlsAwareTransport) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Transport:     tlsAwareHttpTransport,
		CheckRedirect: nil,
		Jar:           jar,
		Timeout:       10 * time.Second,
	}
}

// TlsAwareTransport abstracts HTTP transport to allow API implementations to dynamically
// configure TLS settings during authentication (e.g., adding client certificates) and manage
// proxy configuration.
type TlsAwareTransport interface {
	http.RoundTripper

	// GetTlsClientConfig returns the current TLS configuration.
	GetTlsClientConfig() *tls.Config
	// SetTlsClientConfig updates the TLS configuration.
	SetTlsClientConfig(*tls.Config)

	// SetProxy sets the proxy function for HTTP requests.
	SetProxy(func(*http.Request) (*url.URL, error))
	// GetProxy returns the current proxy function.
	GetProxy() func(*http.Request) (*url.URL, error)

	// CloseIdleConnections closes all idle HTTP connections.
	CloseIdleConnections()
}

var _ TlsAwareTransport = (*TlsAwareHttpTransport)(nil)

// TlsAwareHttpTransport is a concrete implementation of TlsAwareTransport that wraps http.Transport.
type TlsAwareHttpTransport struct {
	*http.Transport
}

// NewTlsAwareHttpTransport creates a TlsAwareHttpTransport with default HTTP/2 and TLS settings.
func NewTlsAwareHttpTransport(cfg *ComponentsConfig) *TlsAwareHttpTransport {
	tlsClientConfig, _ := rest_util.NewTlsConfig()

	authAwareTransport := &TlsAwareHttpTransport{
		&http.Transport{
			TLSClientConfig:       tlsClientConfig,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          10,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	if cfg != nil && cfg.Proxy != nil {
		authAwareTransport.Proxy = cfg.Proxy
	}

	return authAwareTransport
}

// GetProxy returns the proxy function currently set on the transport.
func (a *TlsAwareHttpTransport) GetProxy() func(*http.Request) (*url.URL, error) {
	return a.Proxy
}

// SetProxy sets the proxy function for the transport.
func (a *TlsAwareHttpTransport) SetProxy(proxyFunc func(*http.Request) (*url.URL, error)) {
	a.Proxy = proxyFunc
}

// GetTlsClientConfig returns the TLS configuration from the underlying transport.
func (a *TlsAwareHttpTransport) GetTlsClientConfig() *tls.Config {
	return a.TLSClientConfig
}

// SetTlsClientConfig updates the TLS configuration on the underlying transport.
func (a *TlsAwareHttpTransport) SetTlsClientConfig(config *tls.Config) {
	a.TLSClientConfig = config
}
