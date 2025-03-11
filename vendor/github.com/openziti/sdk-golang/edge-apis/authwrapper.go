// Package edge_apis_2 edge_apis_2 provides a wrapper around the generated Edge Client and Management APIs improve ease
// of use.
package edge_apis

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/edge-api/rest_client_api_client"
	clientAuth "github.com/openziti/edge-api/rest_client_api_client/authentication"
	clientControllers "github.com/openziti/edge-api/rest_client_api_client/controllers"
	clientApiSession "github.com/openziti/edge-api/rest_client_api_client/current_api_session"
	clientInfo "github.com/openziti/edge-api/rest_client_api_client/informational"
	"github.com/openziti/edge-api/rest_management_api_client"
	manAuth "github.com/openziti/edge-api/rest_management_api_client/authentication"
	manControllers "github.com/openziti/edge-api/rest_management_api_client/controllers"
	manCurApiSession "github.com/openziti/edge-api/rest_management_api_client/current_api_session"
	manInfo "github.com/openziti/edge-api/rest_management_api_client/informational"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/edge-api/rest_util"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/foundation/v2/stringz"
	"github.com/pkg/errors"
	"github.com/zitadel/oidc/v2/pkg/client/tokenexchange"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	AuthRequestIdHeader = "auth-request-id"
	TotpRequiredHeader  = "totp-required"
)

// AuthEnabledApi is used as a sentinel interface to detect APIs that support authentication and to work around a golang
// limitation dealing with accessing field of generically typed fields.
type AuthEnabledApi interface {
	//Authenticate will attempt to issue an authentication request using the provided credentials and http client.
	//These functions act as abstraction around the underlying go-swagger generated client and will use the default
	//http client if not provided.
	Authenticate(credentials Credentials, configTypes []string, httpClient *http.Client) (ApiSession, error)
	SetUseOidc(bool)
	ListControllers() (*rest_model.ControllersList, error)
	GetClientTransportPool() ClientTransportPool
	SetClientTransportPool(ClientTransportPool)
}

type ApiSession interface {
	//GetAccessHeader returns the HTTP header name and value that should be used to represent this ApiSession
	GetAccessHeader() (string, string)

	//AuthenticateRequest fulfills the interface defined by the OpenAPI libraries to authenticate client HTTP requests
	AuthenticateRequest(request runtime.ClientRequest, _ strfmt.Registry) error

	//GetToken returns the ApiSessions' token bytes
	GetToken() []byte

	//GetExpiresAt returns the time when the ApiSession will expire.
	GetExpiresAt() *time.Time

	//GetAuthQueries returns a list of authentication queries the ApiSession is subjected to
	GetAuthQueries() rest_model.AuthQueryList

	//GetIdentityName returns the name of the authenticating identity
	GetIdentityName() string

	//GetIdentityId returns the id of the authenticating identity
	GetIdentityId() string

	//GetId returns the id of the ApiSession
	GetId() string

	//RequiresRouterTokenUpdate returns true if the token is a bearer token requires updating on edge router connections.
	RequiresRouterTokenUpdate() bool

	GetRequestHeaders() http.Header
}

var _ ApiSession = (*ApiSessionLegacy)(nil)
var _ ApiSession = (*ApiSessionOidc)(nil)

// ApiSessionLegacy represents OpenZiti's original authentication API Session Detail, supplied in the `zt-session` header.
// It has been supplanted by OIDC authentication represented by ApiSessionOidc.
type ApiSessionLegacy struct {
	Detail         *rest_model.CurrentAPISessionDetail
	RequestHeaders http.Header
}

func (a *ApiSessionLegacy) GetRequestHeaders() http.Header {
	return a.RequestHeaders
}

func (a *ApiSessionLegacy) RequiresRouterTokenUpdate() bool {
	return false
}

func (a *ApiSessionLegacy) GetId() string {
	return stringz.OrEmpty(a.Detail.ID)
}

func (a *ApiSessionLegacy) GetIdentityName() string {
	return a.Detail.Identity.Name
}

func (a *ApiSessionLegacy) GetIdentityId() string {
	return stringz.OrEmpty(a.Detail.IdentityID)
}

// GetAccessHeader returns the header and header token value should be used for authentication requests
func (a *ApiSessionLegacy) GetAccessHeader() (string, string) {
	if a.Detail != nil && a.Detail.Token != nil {
		return "zt-session", *a.Detail.Token
	}

	return "", ""
}

func (a *ApiSessionLegacy) AuthenticateRequest(request runtime.ClientRequest, _ strfmt.Registry) error {
	if a == nil {
		return errors.New("api session is nil")
	}

	for h, v := range a.RequestHeaders {
		err := request.SetHeaderParam(h, v...)
		if err != nil {
			return err
		}
	}

	//legacy does not support multiple zt-session headers, so we can it sfely
	header, val := a.GetAccessHeader()
	err := request.SetHeaderParam(header, val)
	if err != nil {
		return err
	}

	return nil
}

func (a *ApiSessionLegacy) GetToken() []byte {
	if a.Detail != nil && a.Detail.Token != nil {
		return []byte(*a.Detail.Token)
	}

	return nil
}

func (a *ApiSessionLegacy) GetAuthQueries() rest_model.AuthQueryList {
	return a.Detail.AuthQueries
}

func (a *ApiSessionLegacy) GetExpiresAt() *time.Time {
	if a.Detail != nil {
		return (*time.Time)(a.Detail.ExpiresAt)
	}

	return nil
}

// ApiSessionOidc represents an authenticated session backed by OIDC tokens.
type ApiSessionOidc struct {
	OidcTokens     *oidc.Tokens[*oidc.IDTokenClaims]
	RequestHeaders http.Header
}

func (a *ApiSessionOidc) GetRequestHeaders() http.Header {
	return a.RequestHeaders
}

func (a *ApiSessionOidc) RequiresRouterTokenUpdate() bool {
	return true
}

func (a *ApiSessionOidc) GetAccessClaims() (*ApiAccessClaims, error) {
	claims := &ApiAccessClaims{}

	parser := jwt.NewParser()
	_, _, err := parser.ParseUnverified(a.OidcTokens.AccessToken, claims)

	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (a *ApiSessionOidc) GetId() string {
	claims, err := a.GetAccessClaims()

	if err != nil {
		return ""
	}

	return claims.ApiSessionId
}

func (a *ApiSessionOidc) GetIdentityName() string {
	return a.OidcTokens.IDTokenClaims.Name
}

func (a *ApiSessionOidc) GetIdentityId() string {
	return a.OidcTokens.IDTokenClaims.Subject
}

// GetAccessHeader returns the header and header token value should be used for authentication requests
func (a *ApiSessionOidc) GetAccessHeader() (string, string) {
	if a.OidcTokens != nil {
		return "authorization", "Bearer " + a.OidcTokens.AccessToken
	}

	return "", ""
}

func (a *ApiSessionOidc) AuthenticateRequest(request runtime.ClientRequest, _ strfmt.Registry) error {
	if a == nil {
		return errors.New("api session is nil")
	}

	if a.RequestHeaders == nil {
		a.RequestHeaders = http.Header{}
	}

	//multiple Authorization headers are allowed, obtain all auth header candidates
	primaryAuthHeader, primaryAuthValue := a.GetAccessHeader()
	altAuthValues := a.RequestHeaders.Get(primaryAuthHeader)

	authValues := []string{primaryAuthValue}

	if len(altAuthValues) > 0 {
		authValues = append(authValues, altAuthValues)
	}

	//set request headers
	for h, v := range a.RequestHeaders {
		err := request.SetHeaderParam(h, v...)
		if err != nil {
			return err
		}
	}

	//restore auth headers
	err := request.SetHeaderParam(primaryAuthHeader, authValues...)

	if err != nil {
		return err
	}

	return nil
}

func (a *ApiSessionOidc) GetToken() []byte {
	if a.OidcTokens != nil && a.OidcTokens.AccessToken != "" {
		return []byte(a.OidcTokens.AccessToken)
	}

	return nil
}

func (a *ApiSessionOidc) GetAuthQueries() rest_model.AuthQueryList {
	//todo convert JWT auth queries to rest_model.AuthQueryList
	return nil
}

func (a *ApiSessionOidc) GetExpiresAt() *time.Time {
	if a.OidcTokens != nil {
		return &a.OidcTokens.Expiry
	}
	return nil
}

var _ AuthEnabledApi = (*ZitiEdgeManagement)(nil)

// ZitiEdgeManagement is an alias of the go-swagger generated client that allows this package to add additional
// functionality to the alias type to implement the AuthEnabledApi interface.
type ZitiEdgeManagement struct {
	*rest_management_api_client.ZitiEdgeManagement
	// useOidc tracks if OIDC auth should be used
	useOidc bool

	// useOidcExplicitlySet signals if useOidc was set from an external caller and should be used as is
	useOidcExplicitlySet bool

	// oidcDynamicallyEnabled will cause the client to check the controller for OIDC support and use if possible as long as useOidc was not explicitly set
	oidcDynamicallyEnabled bool //currently defaults false till HA release

	versionOnce sync.Once
	versionInfo *rest_model.Version

	TotpCallback        func(chan string)
	ClientTransportPool ClientTransportPool
}

func (self *ZitiEdgeManagement) SetClientTransportPool(transportPool ClientTransportPool) {
	self.ClientTransportPool = transportPool
}

func (self *ZitiEdgeManagement) GetClientTransportPool() ClientTransportPool {
	return self.ClientTransportPool
}

func (self *ZitiEdgeManagement) ListControllers() (*rest_model.ControllersList, error) {
	params := manControllers.NewListControllersParams()
	resp, err := self.Controllers.ListControllers(params, nil)
	if err != nil {
		return nil, err
	}

	return &resp.GetPayload().Data, nil
}

func (self *ZitiEdgeManagement) Authenticate(credentials Credentials, configTypes []string, httpClient *http.Client) (ApiSession, error) {
	self.versionOnce.Do(func() {
		if self.useOidcExplicitlySet {
			return
		}

		if self.oidcDynamicallyEnabled {
			versionParams := manInfo.NewListVersionParams()

			versionResp, _ := self.Informational.ListVersion(versionParams)

			if versionResp != nil {
				self.versionInfo = versionResp.Payload.Data
				self.useOidc = stringz.Contains(self.versionInfo.Capabilities, string(rest_model.CapabilitiesOIDCAUTH))
			}
		} else {
			self.useOidc = false
		}
	})

	if self.useOidc {
		return self.oidcAuth(credentials, configTypes, httpClient)
	}

	return self.legacyAuth(credentials, configTypes, httpClient)
}

func (self *ZitiEdgeManagement) legacyAuth(credentials Credentials, configTypes []string, httpClient *http.Client) (ApiSession, error) {
	params := manAuth.NewAuthenticateParams()
	params.Auth = credentials.Payload()
	params.Method = credentials.Method()
	params.Auth.ConfigTypes = append(params.Auth.ConfigTypes, configTypes...)

	certs := credentials.TlsCerts()
	if len(certs) != 0 {
		if transport, ok := httpClient.Transport.(*http.Transport); ok {
			transport.TLSClientConfig.Certificates = certs
			transport.CloseIdleConnections()
		}
	}

	resp, err := self.Authentication.Authenticate(params, getClientAuthInfoOp(credentials, httpClient))

	if err != nil {
		return nil, err
	}

	return &ApiSessionLegacy{
		Detail:         resp.GetPayload().Data,
		RequestHeaders: credentials.GetRequestHeaders()}, err
}

func (self *ZitiEdgeManagement) oidcAuth(credentials Credentials, configTypeOverrides []string, httpClient *http.Client) (ApiSession, error) {
	return oidcAuth(self.ClientTransportPool, credentials, configTypeOverrides, httpClient, self.TotpCallback)
}

func (self *ZitiEdgeManagement) SetUseOidc(use bool) {
	self.useOidcExplicitlySet = true
	self.useOidc = use
}

func (self *ZitiEdgeManagement) SetAllowOidcDynamicallyEnabled(allow bool) {
	self.oidcDynamicallyEnabled = allow
}

func (self *ZitiEdgeManagement) RefreshApiSession(apiSession ApiSession, httpClient *http.Client) (ApiSession, error) {
	switch s := apiSession.(type) {
	case *ApiSessionLegacy:
		params := manCurApiSession.NewGetCurrentAPISessionParams()
		_, err := self.CurrentAPISession.GetCurrentAPISession(params, s)

		if err != nil {
			return nil, rest_util.WrapErr(err)
		}

		return s, nil
	case *ApiSessionOidc:
		tokens, err := self.ExchangeTokens(s.OidcTokens, httpClient)

		if err != nil {
			return nil, err
		}

		return &ApiSessionOidc{
			OidcTokens:     tokens,
			RequestHeaders: apiSession.GetRequestHeaders(),
		}, nil
	}

	return nil, errors.New("api session does not have any tokens")
}

func (self *ZitiEdgeManagement) ExchangeTokens(curTokens *oidc.Tokens[*oidc.IDTokenClaims], httpClient *http.Client) (*oidc.Tokens[*oidc.IDTokenClaims], error) {
	return exchangeTokens(self.ClientTransportPool, curTokens, httpClient)
}

var _ AuthEnabledApi = (*ZitiEdgeClient)(nil)

// ZitiEdgeClient is an alias of the go-swagger generated client that allows this package to add additional
// functionality to the alias type to implement the AuthEnabledApi interface.
type ZitiEdgeClient struct {
	*rest_client_api_client.ZitiEdgeClient
	// useOidc tracks if OIDC auth should be used
	useOidc bool

	// useOidcExplicitlySet signals if useOidc was set from an external caller and should be used as is
	useOidcExplicitlySet bool

	// oidcDynamicallyEnabled will cause the client to check the controller for OIDC support and use if possible as long as useOidc was not explicitly set.
	oidcDynamicallyEnabled bool //currently defaults false till HA release

	versionInfo *rest_model.Version
	versionOnce sync.Once

	TotpCallback        func(chan string)
	ClientTransportPool ClientTransportPool
}

func (self *ZitiEdgeClient) GetClientTransportPool() ClientTransportPool {
	return self.ClientTransportPool
}

func (self *ZitiEdgeClient) SetClientTransportPool(transportPool ClientTransportPool) {
	self.ClientTransportPool = transportPool
}

func (self *ZitiEdgeClient) ListControllers() (*rest_model.ControllersList, error) {
	params := clientControllers.NewListControllersParams()
	resp, err := self.Controllers.ListControllers(params, nil)
	if err != nil {
		return nil, err
	}

	return &resp.GetPayload().Data, nil
}

func (self *ZitiEdgeClient) Authenticate(credentials Credentials, configTypesOverrides []string, httpClient *http.Client) (ApiSession, error) {
	self.versionOnce.Do(func() {
		if self.useOidcExplicitlySet {
			return
		}

		if self.oidcDynamicallyEnabled {
			versionParams := clientInfo.NewListVersionParams()

			versionResp, _ := self.Informational.ListVersion(versionParams)

			if versionResp != nil {
				self.versionInfo = versionResp.Payload.Data
				self.useOidc = stringz.Contains(self.versionInfo.Capabilities, string(rest_model.CapabilitiesOIDCAUTH))
			}
		} else {
			self.useOidc = false
		}
	})

	if self.useOidc {
		return self.oidcAuth(credentials, configTypesOverrides, httpClient)
	}

	return self.legacyAuth(credentials, configTypesOverrides, httpClient)
}

func (self *ZitiEdgeClient) legacyAuth(credentials Credentials, configTypes []string, httpClient *http.Client) (ApiSession, error) {
	params := clientAuth.NewAuthenticateParams()
	params.Auth = credentials.Payload()
	params.Method = credentials.Method()
	params.Auth.ConfigTypes = append(params.Auth.ConfigTypes, configTypes...)

	certs := credentials.TlsCerts()
	if len(certs) != 0 {
		if transport, ok := httpClient.Transport.(*http.Transport); ok {
			transport.TLSClientConfig.Certificates = certs
			transport.CloseIdleConnections()
		}
	}

	resp, err := self.Authentication.Authenticate(params, getClientAuthInfoOp(credentials, httpClient))

	if err != nil {
		return nil, err
	}

	return &ApiSessionLegacy{Detail: resp.GetPayload().Data, RequestHeaders: credentials.GetRequestHeaders()}, err
}

func (self *ZitiEdgeClient) oidcAuth(credentials Credentials, configTypeOverrides []string, httpClient *http.Client) (ApiSession, error) {
	return oidcAuth(self.ClientTransportPool, credentials, configTypeOverrides, httpClient, self.TotpCallback)
}

func (self *ZitiEdgeClient) SetUseOidc(use bool) {
	self.useOidcExplicitlySet = true
	self.useOidc = use
}

func (self *ZitiEdgeClient) SetAllowOidcDynamicallyEnabled(allow bool) {
	self.oidcDynamicallyEnabled = allow
}

func (self *ZitiEdgeClient) RefreshApiSession(apiSession ApiSession, httpClient *http.Client) (ApiSession, error) {
	switch s := apiSession.(type) {
	case *ApiSessionLegacy:
		params := clientApiSession.NewGetCurrentAPISessionParams()
		newApiSessionDetail, err := self.CurrentAPISession.GetCurrentAPISession(params, s)

		if err != nil {
			return nil, rest_util.WrapErr(err)
		}

		newApiSession := &ApiSessionLegacy{
			Detail:         newApiSessionDetail.Payload.Data,
			RequestHeaders: apiSession.GetRequestHeaders(),
		}

		return newApiSession, nil
	case *ApiSessionOidc:
		tokens, err := self.ExchangeTokens(s.OidcTokens, httpClient)

		if err != nil {
			return nil, err
		}

		return &ApiSessionOidc{
			OidcTokens:     tokens,
			RequestHeaders: apiSession.GetRequestHeaders(),
		}, nil
	}

	return nil, errors.New("api session does not have any tokens")
}

func (self *ZitiEdgeClient) ExchangeTokens(curTokens *oidc.Tokens[*oidc.IDTokenClaims], httpClient *http.Client) (*oidc.Tokens[*oidc.IDTokenClaims], error) {
	return exchangeTokens(self.ClientTransportPool, curTokens, httpClient)
}

func exchangeTokens(clientTransportPool ClientTransportPool, curTokens *oidc.Tokens[*oidc.IDTokenClaims], client *http.Client) (*oidc.Tokens[*oidc.IDTokenClaims], error) {
	subjectToken := curTokens.RefreshToken
	subjectTokenType := oidc.RefreshTokenType

	// if subjectToken is "", then we don't have a refresh token, attempt to exchange a non-expired access token
	if subjectToken == "" {
		if curTokens.Expiry.Before(time.Now()) {
			return nil, errors.New("cannot exchange token: refresh token not found, access token expired")
		}

		if curTokens.AccessToken == "" {
			return nil, errors.New("cannot exchange token: refresh token not found, access token not found")
		}
		subjectToken = curTokens.AccessToken
		subjectTokenType = oidc.AccessTokenType
	}

	var outTokens *oidc.Tokens[*oidc.IDTokenClaims]

	_, err := clientTransportPool.TryTransportForF(func(transport *ApiClientTransport) (any, error) {
		apiHost := transport.ApiUrl.Host
		issuer := "https://" + apiHost + "/oidc"
		tokenEndpoint := "https://" + apiHost + "/oidc/oauth/token"

		te, err := tokenexchange.NewTokenExchangerClientCredentials(issuer, "native", "", tokenexchange.WithHTTPClient(client), tokenexchange.WithStaticTokenEndpoint(issuer, tokenEndpoint))

		if err != nil {
			return nil, err
		}

		var tokenResponse *oidc.TokenExchangeResponse

		now := time.Now()

		switch subjectTokenType {
		case oidc.RefreshTokenType:
			tokenResponse, err = tokenexchange.ExchangeToken(te, subjectToken, subjectTokenType, "", "", nil, nil, nil, oidc.RefreshTokenType)
		case oidc.AccessTokenType:
			tokenResponse, err = tokenexchange.ExchangeToken(te, subjectToken, subjectTokenType, "", "", nil, nil, nil, oidc.AccessTokenType)
		}

		if err != nil {
			return nil, err
		}

		idResp, err := tokenexchange.ExchangeToken(te, subjectToken, subjectTokenType, "", "", nil, nil, nil, oidc.IDTokenType)

		if err != nil {
			return nil, err
		}

		idClaims := &IdClaims{}

		//access token is used to hold id token per zitadel comments
		_, _, err = jwt.NewParser().ParseUnverified(idResp.AccessToken, idClaims)

		if err != nil {
			return nil, err
		}

		outTokens = &oidc.Tokens[*oidc.IDTokenClaims]{
			Token: &oauth2.Token{
				AccessToken:  tokenResponse.AccessToken,
				TokenType:    tokenResponse.TokenType,
				RefreshToken: tokenResponse.RefreshToken,
				Expiry:       now.Add(time.Second * time.Duration(tokenResponse.ExpiresIn)),
			},
			IDTokenClaims: &idClaims.IDTokenClaims,
			IDToken:       idResp.AccessToken, //access token field is used to hold id token per zitadel comments
		}

		return outTokens, nil
	})

	if err != nil {
		return nil, err
	}

	return outTokens, nil
}

type authPayload struct {
	*rest_model.Authenticate
	AuthRequestId string `json:"id"`
}

type totpCodePayload struct {
	rest_model.MfaCode
	AuthRequestId string `json:"id"`
}

func (a *authPayload) toValues() url.Values {
	result := url.Values{
		"id":            []string{a.AuthRequestId},
		"password":      []string{string(a.Password)},
		"username":      []string{string(a.Username)},
		"configTypes":   a.ConfigTypes,
		"envArch":       []string{a.EnvInfo.Arch},
		"envOs":         []string{a.EnvInfo.Os},
		"envOsRelease":  []string{a.EnvInfo.OsRelease},
		"envOsVersion":  []string{a.EnvInfo.OsVersion},
		"sdkAppID":      []string{a.SdkInfo.AppID},
		"sdkAppVersion": []string{a.SdkInfo.AppVersion},
		"sdkBranch":     []string{a.SdkInfo.Branch},
		"sdkRevision":   []string{a.SdkInfo.Revision},
		"sdkType":       []string{a.SdkInfo.Type},
		"sdkVersion":    []string{a.SdkInfo.Version},
	}

	return result
}

func oidcAuth(clientTransportPool ClientTransportPool, credentials Credentials, configTypeOverrides []string, httpClient *http.Client, totpCallback func(chan string)) (ApiSession, error) {
	payload := &authPayload{
		Authenticate: credentials.Payload(),
	}
	method := credentials.Method()

	if configTypeOverrides != nil {
		payload.ConfigTypes = configTypeOverrides
	}

	certs := credentials.TlsCerts()

	if len(certs) != 0 {
		if transport, ok := httpClient.Transport.(*http.Transport); ok {
			transport.TLSClientConfig.Certificates = certs
			transport.CloseIdleConnections()
		}
	}

	var outTokens *oidc.Tokens[*oidc.IDTokenClaims]

	_, err := clientTransportPool.TryTransportForF(func(transport *ApiClientTransport) (any, error) {
		rpServer, err := newLocalRpServer(transport.ApiUrl.Host, method)

		if err != nil {
			return nil, err
		}

		rpServer.Start()
		defer rpServer.Stop()

		client := resty.NewWithClient(httpClient)
		apiHost := transport.ApiUrl.Hostname()

		client.SetRedirectPolicy(resty.DomainCheckRedirectPolicy("127.0.0.1", "localhost", apiHost))
		resp, err := client.R().Get(rpServer.LoginUri)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("local rp login response is expected to be HTTP status %d got %d with body: %s", http.StatusOK, resp.StatusCode(), resp.Body())
		}
		payload.AuthRequestId = resp.Header().Get(AuthRequestIdHeader)

		if payload.AuthRequestId == "" {
			return nil, errors.New("could not find auth request id header")
		}

		opLoginUri := "https://" + resp.RawResponse.Request.URL.Host + "/oidc/login/" + method
		totpUri := "https://" + resp.RawResponse.Request.URL.Host + "/oidc/login/totp"

		formData := payload.toValues()

		req := client.R()
		clientRequest := asClientRequest(req, client)

		err = credentials.AuthenticateRequest(clientRequest, strfmt.Default)

		if err != nil {
			return nil, err
		}

		resp, err = req.SetFormDataFromValues(formData).Post(opLoginUri)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("remote op login response is expected to be HTTP status %d got %d with body: %s", http.StatusOK, resp.StatusCode(), resp.Body())
		}

		authRequestId := payload.AuthRequestId
		totpRequiredHeader := resp.Header().Get(TotpRequiredHeader)
		totpRequired := totpRequiredHeader != ""
		totpCode := ""

		if totpRequired {

			if totpCallback == nil {
				return nil, errors.New("totp is required but not totp callback was defined")
			}
			codeChan := make(chan string)
			go totpCallback(codeChan)

			select {
			case code := <-codeChan:
				totpCode = code
			case <-time.After(30 * time.Minute):
				return nil, fmt.Errorf("timedout waiting for totpT callback")
			}

			resp, err = client.R().SetBody(&totpCodePayload{
				MfaCode: rest_model.MfaCode{
					Code: &totpCode,
				},
				AuthRequestId: authRequestId,
			}).Post(totpUri)

			if err != nil {
				return nil, err
			}

			if resp.StatusCode() != http.StatusOK {
				apiErr := &errorz.ApiError{}
				err = json.Unmarshal(resp.Body(), apiErr)

				if err != nil {
					return nil, fmt.Errorf("could not verify TOTP MFA code recieved %d - could not parse body: %s", resp.StatusCode(), string(resp.Body()))
				}

				return nil, apiErr
			}
		}

		var tokens *oidc.Tokens[*oidc.IDTokenClaims]
		select {
		case tokens = <-rpServer.TokenChan:
		case <-time.After(30 * time.Minute):
		}

		if tokens == nil {
			return nil, errors.New("authentication did not complete, received nil tokens")
		}
		outTokens = tokens

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return &ApiSessionOidc{
		OidcTokens:     outTokens,
		RequestHeaders: credentials.GetRequestHeaders(),
	}, nil
}

// restyClientRequest is meant to mimic open api's client request which is a combination
// of resty's request and client.
type restyClientRequest struct {
	restyRequest *resty.Request
	restyClient  *resty.Client
}

func (r *restyClientRequest) SetHeaderParam(s string, s2 ...string) error {
	r.restyRequest.Header[s] = s2
	return nil
}

func (r *restyClientRequest) GetHeaderParams() http.Header {
	return r.restyRequest.Header
}

func (r *restyClientRequest) SetQueryParam(s string, s2 ...string) error {
	r.restyRequest.QueryParam[s] = s2
	return nil
}

func (r *restyClientRequest) SetFormParam(s string, s2 ...string) error {
	r.restyRequest.FormData[s] = s2
	return nil
}

func (r *restyClientRequest) SetPathParam(s string, s2 string) error {
	r.restyRequest.PathParams[s] = s2
	return nil
}

func (r *restyClientRequest) GetQueryParams() url.Values {
	return r.restyRequest.QueryParam
}

func (r *restyClientRequest) SetFileParam(s string, closer ...runtime.NamedReadCloser) error {
	for _, curCloser := range closer {
		r.restyRequest.SetFileReader(s, curCloser.Name(), curCloser)
	}

	return nil
}

func (r *restyClientRequest) SetBodyParam(i interface{}) error {
	r.restyRequest.SetBody(i)
	return nil
}

func (r *restyClientRequest) SetTimeout(duration time.Duration) error {
	r.restyClient.SetTimeout(duration)
	return nil
}

func (r *restyClientRequest) GetMethod() string {
	return r.restyRequest.Method
}

func (r *restyClientRequest) GetPath() string {
	return r.restyRequest.URL
}

func (r *restyClientRequest) GetBody() []byte {
	return r.restyRequest.Body.([]byte)
}

func (r *restyClientRequest) GetBodyParam() interface{} {
	return r.restyRequest.Body
}

func (r *restyClientRequest) GetFileParam() map[string][]runtime.NamedReadCloser {
	return nil
}

func asClientRequest(request *resty.Request, client *resty.Client) runtime.ClientRequest {
	return &restyClientRequest{request, client}
}
