package edge_apis

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/identity"
	"github.com/openziti/sdk-golang/ziti/edge/network"
	"github.com/openziti/sdk-golang/ziti/sdkinfo"
	"net/http"
)

// Credentials represents the minimal information needed across all authentication mechanisms to authenticate an identity
// to an OpenZiti network.
type Credentials interface {
	// Payload constructs the objects that represent the JSON authentication payload for this set of credentials.
	Payload() *rest_model.Authenticate

	// TlsCerts returns zero or more tls.Certificates used for client authentication.
	TlsCerts() []tls.Certificate

	// GetCaPool returns the CA pool that this credential was configured to trust.
	GetCaPool() *x509.CertPool

	// Method returns the authentication necessary to complete an authentication request.
	Method() string

	// AddAuthHeader adds a header for all authentication requests.
	AddAuthHeader(key, value string)

	// AddRequestHeader adds a header for all requests after authentication
	AddRequestHeader(key, value string)

	// AddJWT adds additional JWTs to the credentials. Used to satisfy secondary authentication/MFA requirements. The
	// provided token should be the base64 encoded version of the token.
	AddJWT(string)

	// ClientAuthInfoWriter is used to pass a Credentials instance to the openapi runtime to authenticate outgoing
	//requests.
	runtime.ClientAuthInfoWriter

	// GetRequestHeaders returns a set of headers to use after authentication during normal HTTP operations
	GetRequestHeaders() http.Header
}

// IdentityProvider is a sentinel interface used to determine whether the backing Credentials instance can provide
// an Identity that can provide a certificate and private key used to initiate mTLS connections.
type IdentityProvider interface {
	GetIdentity() identity.Identity
}

// toTlsCerts converts an array of certificates into a single tls.Certificate. Index zero is assumed to be the leaf
// certificate and all subsequent certificates to be the support certificate chain that should be sent to servers.
// At least one certificate must be provided.
func toTlsCerts(certs []*x509.Certificate, key crypto.PrivateKey) tls.Certificate {
	tlsCert := tls.Certificate{
		PrivateKey: key,
		Leaf:       certs[0],
	}
	for _, cert := range certs {
		tlsCert.Certificate = append(tlsCert.Certificate, cert.Raw)
	}

	return tlsCert
}

// getClientAuthInfoOp returns a one-off runtime.ClientOperation used to authenticate single requests without altering
// the authentication operation of the entire client runtime.
func getClientAuthInfoOp(credentials Credentials, client *http.Client) func(*runtime.ClientOperation) {
	return func(operation *runtime.ClientOperation) {
		operation.AuthInfo = credentials

		certs := credentials.TlsCerts()

		if len(certs) != 0 {
			operation.Client = client
			if transport, ok := operation.Client.Transport.(*http.Transport); ok {
				transport.TLSClientConfig.Certificates = certs
			}
		}
	}
}

// BaseCredentials is a shared struct of information all Credentials implementations require.
type BaseCredentials struct {
	// ConfigTypes is used to set the configuration types for services during authentication
	ConfigTypes []string

	// AuthHeaders is a map of strings to string arrays of headers to send with auth requests.
	AuthHeaders http.Header

	// RequestHeaders is a map of string to string arrays of headers to send on non-authentication requests.
	RequestHeaders http.Header

	// EnvInfo is provided during authentication to set environmental information about the client.
	EnvInfo *rest_model.EnvInfo

	// SdkInfo is provided during authentication to set SDK information about the client.
	SdkInfo *rest_model.SdkInfo

	// CaPool will override the client's default certificate pool if set to a non-nil value.
	CaPool *x509.CertPool
}

// Payload will produce the object used to construct the body of an authentication requests. The base version
// sets shared information available in BaseCredentials.
func (c *BaseCredentials) Payload() *rest_model.Authenticate {
	envInfo, sdkInfo := sdkinfo.GetSdkInfo()

	if c.EnvInfo != nil {
		envInfo = c.EnvInfo
	}

	if c.SdkInfo != nil {
		sdkInfo = c.SdkInfo
	}

	return &rest_model.Authenticate{
		ConfigTypes: c.ConfigTypes,
		EnvInfo:     envInfo,
		SdkInfo:     sdkInfo,
	}
}

// GetCaPool provides a base implementation to return the certificate pool of a Credentials instance.
func (c *BaseCredentials) GetCaPool() *x509.CertPool {
	return c.CaPool
}

// AddAuthHeader provides a base implementation to add a header to authentication requests.
func (c *BaseCredentials) AddAuthHeader(key, value string) {
	if c.AuthHeaders == nil {
		c.AuthHeaders = http.Header{}
	}
	c.AuthHeaders.Add(key, value)
}

// AddRequestHeader provides a base implementation to add a header to all requests after authentication.
func (c *BaseCredentials) AddRequestHeader(key, value string) {
	if c.RequestHeaders == nil {
		c.RequestHeaders = http.Header{}
	}

	c.RequestHeaders.Add(key, value)
}

// AddJWT adds additional JWTs to the credentials. Used to satisfy secondary authentication/MFA requirements. The
// provided token should be the base64 encoded version of the token. Convenience function for AddHeader.
func (c *BaseCredentials) AddJWT(token string) {
	c.AddAuthHeader("Authorization", "Bearer "+token)
	c.AddRequestHeader("Authorization", "Bearer "+token)
}

// AuthenticateRequest provides a base implementation to authenticate an outgoing request. This is provided here
// for authentication methods such as `cert` which do not have to provide any more request level information.
func (c *BaseCredentials) AuthenticateRequest(request runtime.ClientRequest, _ strfmt.Registry) error {
	var errors []error

	for hName, hVals := range c.AuthHeaders {
		for _, hVal := range hVals {
			err := request.SetHeaderParam(hName, hVal)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return network.MultipleErrors(errors)
	}
	return nil
}

// ProcessRequest proves a base implemmentation mutate runtime.ClientRequests as they are sent out after
// authentication. Useful for adding headers.
func (c *BaseCredentials) ProcessRequest(request runtime.ClientRequest, _ strfmt.Registry) error {
	var errors []error

	for hName, hVals := range c.RequestHeaders {
		for _, hVal := range hVals {
			err := request.SetHeaderParam(hName, hVal)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return network.MultipleErrors(errors)
	}
	return nil
}

// GetRequestHeaders returns headers that should be sent on requests post authentication.
func (c *BaseCredentials) GetRequestHeaders() http.Header {
	return c.RequestHeaders
}

// TlsCerts provides a base implementation of returning the tls.Certificate array that will be used to setup
// mTLS connections. This is provided here for authentication methods that do not initially require mTLS (e.g. JWTs).
func (c *BaseCredentials) TlsCerts() []tls.Certificate {
	return nil
}

var _ Credentials = &CertCredentials{}

// CertCredentials represents authentication using certificates that are not from an Identity configuration file.
type CertCredentials struct {
	BaseCredentials
	Certs []*x509.Certificate
	Key   crypto.PrivateKey
}

// NewCertCredentials creates Credentials instance based upon an array of certificates. At least one certificate must
// be provided and the certificate at index zero is assumed to be the leaf client certificate that pairs with the
// provided private key. All other certificates are assumed to support the leaf client certificate as a chain.
func NewCertCredentials(certs []*x509.Certificate, key crypto.PrivateKey) *CertCredentials {
	return &CertCredentials{
		BaseCredentials: BaseCredentials{},
		Certs:           certs,
		Key:             key,
	}
}

func (c *CertCredentials) Method() string {
	return "cert"
}

func (c *CertCredentials) TlsCerts() []tls.Certificate {
	return []tls.Certificate{toTlsCerts(c.Certs, c.Key)}
}

func (c *CertCredentials) GetIdentity() identity.Identity {
	return identity.NewClientTokenIdentityWithPool(c.Certs, c.Key, c.GetCaPool())
}

var _ Credentials = &IdentityCredentials{}

type IdentityCredentials struct {
	BaseCredentials
	Identity identity.Identity
}

// NewIdentityCredentials creates a Credentials instance based upon and Identity.
func NewIdentityCredentials(identity identity.Identity) *IdentityCredentials {
	return &IdentityCredentials{
		BaseCredentials: BaseCredentials{},
		Identity:        identity,
	}
}

// NewIdentityCredentialsFromConfig creates a Credentials instance based upon and Identity configuration.
func NewIdentityCredentialsFromConfig(config identity.Config) *IdentityCredentials {
	return &IdentityCredentials{
		BaseCredentials: BaseCredentials{},
		Identity:        &identity.LazyIdentity{Config: &config},
	}
}

func (c *IdentityCredentials) GetIdentity() identity.Identity {
	return c.Identity
}

func (c *IdentityCredentials) Method() string {
	return "cert"
}

func (c *IdentityCredentials) GetCaPool() *x509.CertPool {
	return c.Identity.CA()
}

func (c *IdentityCredentials) TlsCerts() []tls.Certificate {
	tlsCert := c.Identity.Cert()

	if tlsCert != nil {
		return []tls.Certificate{*tlsCert}
	}
	return nil
}

func (c *IdentityCredentials) AuthenticateRequest(request runtime.ClientRequest, reg strfmt.Registry) error {
	return c.BaseCredentials.AuthenticateRequest(request, reg)
}

var _ Credentials = &JwtCredentials{}

type JwtCredentials struct {
	BaseCredentials
	JWT                string
	SendOnEveryRequest bool
}

// NewJwtCredentials creates a Credentials instance based on a JWT obtained from an outside system.
func NewJwtCredentials(jwt string) *JwtCredentials {
	return &JwtCredentials{
		BaseCredentials: BaseCredentials{},
		JWT:             jwt,
	}
}

func (c *JwtCredentials) Method() string {
	return "ext-jwt"
}

func (c *JwtCredentials) AuthenticateRequest(request runtime.ClientRequest, reg strfmt.Registry) error {
	var errors []error

	err := c.BaseCredentials.AuthenticateRequest(request, reg)
	if err != nil {
		errors = append(errors, err)
	}
	err = request.SetHeaderParam("Authorization", "Bearer "+c.JWT)
	if err != nil {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return network.MultipleErrors(errors)
	}
	return nil
}

var _ Credentials = &UpdbCredentials{}

type UpdbCredentials struct {
	BaseCredentials
	Username string
	Password string
}

func (c *UpdbCredentials) Method() string {
	return "password"
}

// NewUpdbCredentials creates a Credentials instance based on a username/passwords combination.
func NewUpdbCredentials(username string, password string) *UpdbCredentials {
	return &UpdbCredentials{
		BaseCredentials: BaseCredentials{},
		Username:        username,
		Password:        password,
	}
}

func (c *UpdbCredentials) Payload() *rest_model.Authenticate {
	payload := c.BaseCredentials.Payload()
	payload.Username = rest_model.Username(c.Username)
	payload.Password = rest_model.Password(c.Password)

	return payload
}

func (c *UpdbCredentials) AuthenticateRequest(request runtime.ClientRequest, reg strfmt.Registry) error {
	return c.BaseCredentials.AuthenticateRequest(request, reg)
}
