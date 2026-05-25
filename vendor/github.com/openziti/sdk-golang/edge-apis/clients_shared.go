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
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/edge-api/rest_util"
	"github.com/zitadel/oidc/v3/pkg/client/tokenexchange"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
)

// DefaultOidcRedirectUri is the default redirect URI for the OIDC PKCE flow that satisfies the default OIDC redirects
// for the Ziti Edge OIDC API. It is not an actual server, rather an intercepted redirect URI that is used to extract
// the resulting OIDC tokens.
const DefaultOidcRedirectUri = "http://localhost:8080/auth/callback"

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

	// SetOidcRedirectUri sets the redirect URI for the OIDC PKCE flow. The default value is used if not set.
	// Should only be necessary to call for custom redirect controller configurations.
	SetOidcRedirectUri(redirectUri string)
}

// OidcAuthResponses contains a set of http.Responses that occur during the OIDC flow. Used for inspection and testing.
type OidcAuthResponses struct {

	// InitResponse is the response from the initial OIDC request to obtain an auth request id/context
	InitResponse *resty.Response

	// PrimaryCredentialResponse is the response provided during primary credential authentication, nil if not reached due to OIDC init errors
	PrimaryCredentialResponse *resty.Response

	// TotpEnrollResponse is the response from starting TOTP enrollment, nil if enrollment was not performed
	TotpEnrollResponse *resty.Response

	// TotpEnrollVerifyResponse is the response from verifying the enrollment TOTP code, nil if enrollment was not performed
	TotpEnrollVerifyResponse *resty.Response

	// TotpResponse is the response provided after the TOTP code was provided for an already-enrolled identity, nil if not reached
	TotpResponse *resty.Response

	// RedirectResponse is the response provided after authentication, nil if never reached
	RedirectResponse *resty.Response
}

// OidcAuthorizeResult holds the result of starting an OIDC PKCE authorization flow.
// It provides the auth request ID, a pre-configured resty client for making raw HTTP calls
// to the OP login endpoints, and an exchange function that completes the flow by trading
// an authorization code for OIDC tokens.
type OidcAuthorizeResult struct {

	// AuthRequestId is the auth request identifier returned by the /oidc/authorize endpoint.
	AuthRequestId string

	// Client is a resty client pre-configured with the correct TLS settings and redirect policy
	// for the OIDC flow. Use it to make raw HTTP calls to OP login endpoints.
	Client *resty.Client

	// Exchange completes the OIDC flow by exchanging an authorization code (extracted from the
	// callback redirect Location header) for OIDC tokens.
	Exchange func(code string) (*oidc.Tokens[*oidc.IDTokenClaims], error)
}

// EdgeOidcAuthConfig represents the options necessary to complete an OAuth 2.0 PKCE authentication flow against an
// OpenZiti controller.
type EdgeOidcAuthConfig struct {
	ClientTransportPool    ClientTransportPool
	Credentials            Credentials
	ConfigTypeOverrides    []string
	HttpClient             *http.Client
	TotpCodeProvider       TotpCodeProvider
	TotpEnrollmentProvider TotpEnrollmentProvider
	RedirectUri            string
	ApiHost                string
}

// ApiClientConfig contains configuration options for creating API clients.
type ApiClientConfig struct {
	ApiUrls                []*url.URL
	CaPool                 *x509.CertPool
	TotpCodeProvider       TotpCodeProvider
	TotpEnrollmentProvider TotpEnrollmentProvider
	Components             *Components
	Proxy                  func(r *http.Request) (*url.URL, error)
}

// exchangeTokens exchanges OIDC tokens for refreshed tokens. It uses refresh tokens preferentially,
// falling back to non-expired access tokens if refresh is unavailable.
func exchangeTokens(clientTransportPool ClientTransportPool, curTokens *oidc.Tokens[*oidc.IDTokenClaims], client *http.Client) (*oidc.Tokens[*oidc.IDTokenClaims], error) {
	subjectToken := ""
	var subjectTokenType oidc.TokenType

	if curTokens.RefreshToken != "" {
		subjectToken = curTokens.RefreshToken
		subjectTokenType = oidc.RefreshTokenType
	} else if curTokens.AccessToken != "" {
		// if subjectToken is "", then we don't have a refresh token, attempt to exchange a non-expired access token
		expired, err := isAccessTokenExpired(curTokens)

		if err != nil {
			return nil, err
		}

		if expired {
			return nil, errors.New("cannot exchange token: refresh token not found, access token expired")
		}

		if curTokens.AccessToken == "" {
			return nil, errors.New("cannot exchange token: refresh token not found, access token not found")
		}
		subjectToken = curTokens.AccessToken
		subjectTokenType = oidc.AccessTokenType
	}

	if subjectToken == "" {
		return nil, errors.New("cannot exchange token: refresh token not found, access token not found or expired")
	}

	var outTokens *oidc.Tokens[*oidc.IDTokenClaims]

	_, err := clientTransportPool.TryTransportForF(func(transport *ApiClientTransport) (any, error) {
		timeoutCtx, cancelF := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelF()

		apiHost := transport.ApiUrl.Host
		issuer := "https://" + apiHost + "/oidc"
		tokenEndpoint := "https://" + apiHost + "/oidc/oauth/token"

		te, err := tokenexchange.NewTokenExchangerClientCredentials(timeoutCtx, issuer, "native", "", tokenexchange.WithHTTPClient(client), tokenexchange.WithStaticTokenEndpoint(issuer, tokenEndpoint))

		if err != nil {
			return nil, err
		}

		var tokenResponse *oidc.TokenExchangeResponse

		now := time.Now()

		switch subjectTokenType {
		case oidc.RefreshTokenType:
			tokenResponse, err = tokenexchange.ExchangeToken(timeoutCtx, te, subjectToken, subjectTokenType, "", "", nil, nil, nil, oidc.RefreshTokenType)
		case oidc.AccessTokenType:
			tokenResponse, err = tokenexchange.ExchangeToken(timeoutCtx, te, subjectToken, subjectTokenType, "", "", nil, nil, nil, oidc.AccessTokenType)
		}

		if err != nil {
			return nil, err
		}

		idResp, err := tokenexchange.ExchangeToken(timeoutCtx, te, subjectToken, subjectTokenType, "", "", nil, nil, nil, oidc.IDTokenType)

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

// isAccessTokenExpired checks if an access token is expired. If token metadata is unavailable,
// it parses the JWT claims to determine expiration.
func isAccessTokenExpired(tokens *oidc.Tokens[*oidc.IDTokenClaims]) (bool, error) {
	if tokens.Expiry.IsZero() {
		//meta data isn't set, we need to parse the token
		idClaims := &IdClaims{}
		_, _, err := jwt.NewParser().ParseUnverified(tokens.AccessToken, idClaims)

		if err != nil {
			return true, fmt.Errorf("token meta data is empty, could not parse token to determine token validity: %w", err)
		}

		//failed to parse out a required exp field for oAuth2, we have no idea of this token is good
		if idClaims.GetExpiration().IsZero() {
			return true, errors.New("token meta data is empty, parsed token does not have an expiration value")
		}

		return idClaims.GetExpiration().Before(time.Now()), nil
	}

	return tokens.Expiry.Before(time.Now()), nil
}

type authPayload struct {
	*rest_model.Authenticate
	AuthRequestId string `json:"id"`
}

type totpCodePayload struct {
	rest_model.MfaCode
	AuthRequestId string `json:"id"`
}

type authRequestIdPayload struct {
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

// oidcAuth performs OIDC authentication using OAuth flow with PKCE.
// It handles TOTP if required and returns an OIDC session with tokens.
func oidcAuth(config *EdgeOidcAuthConfig) (ApiSession, error) {
	if config.Credentials.Method() == AuthMethodEmpty {
		return nil, fmt.Errorf("auth method %s cannot be used for authentication, please provide alternate credentials", AuthMethodEmpty)
	}

	certificates := config.Credentials.TlsCerts()

	if len(certificates) != 0 {
		if transport, ok := config.HttpClient.Transport.(TlsAwareTransport); ok {
			tlsClientConf := transport.GetTlsClientConfig()
			tlsClientConf.Certificates = certificates
			transport.CloseIdleConnections()
		}
	}

	var outTokens *oidc.Tokens[*oidc.IDTokenClaims]

	_, err := config.ClientTransportPool.TryTransportForF(func(transport *ApiClientTransport) (any, error) {
		config.ApiHost = transport.ApiUrl.Host
		edgeOidcAuth := NewEdgeOidcAuthenticator(config)

		var err error
		outTokens, err = edgeOidcAuth.Authenticate()

		if err != nil {
			return nil, err
		}

		return outTokens, nil
	})

	if err != nil {
		return nil, err
	}

	return &ApiSessionOidc{
		OidcTokens:     outTokens,
		RequestHeaders: config.Credentials.GetRequestHeaders(),
	}, nil
}

// EdgeOidcAuthenticator handles the OAuth 2.0 PKCE authentication flow for the Ziti Edge API.
// It submits user credentials to the authorization endpoint, handles optional TOTP verification,
// and exchanges the authorization code for OIDC tokens. The HTTP client follows redirects
// during the authorization flow and extracts the authorization code from the final redirect.
type EdgeOidcAuthenticator struct {
	*EdgeOidcAuthConfig
	restyClient *resty.Client
}

// NewEdgeOidcAuthenticator creates a new EdgeOidcAuthenticator configured for PKCE authentication.
// It sets up an HTTP client with a custom redirect policy that follows redirects during the
// authorization flow but stops when the callback redirect URI is reached, allowing code extraction
// from the redirect URL. The redirectUri parameter defines where the authorization server will
// redirect with the authorization code in the query parameters.
func NewEdgeOidcAuthenticator(config *EdgeOidcAuthConfig) *EdgeOidcAuthenticator {
	client := resty.NewWithClient(config.HttpClient)

	if config.RedirectUri == "" {
		config.RedirectUri = DefaultOidcRedirectUri
	}

	// allows resty to follow redirects for us during the OAuth flow, but not for the end PKCE callback
	// there is no server running for that redirect to hit, as it is this code
	client.SetRedirectPolicy(RedirectUntilUrlPrefix(DefaultOidcRedirectUri))

	return &EdgeOidcAuthenticator{
		EdgeOidcAuthConfig: config,
		restyClient:        client,
	}
}

// SetRedirectUri sets the redirect URI for the authorization server. The default value is
// included in the default Edge OIDC controller configuration, but if it has been set to custom
// values, this function can be used to reflect that configuration.
func (e *EdgeOidcAuthenticator) SetRedirectUri(redirectUri string) {
	e.RedirectUri = redirectUri
}

// Authenticate performs the complete OAuth 2.0 PKCE authentication flow. It initiates authorization
// with PKCE parameters, submits credentials and handles optional TOTP verification, then exchanges
// the resulting authorization code for OIDC tokens.
func (e *EdgeOidcAuthenticator) Authenticate() (*oidc.Tokens[*oidc.IDTokenClaims], error) {
	tokens, _, err := e.AuthenticateWithResponses()
	return tokens, err
}

// AuthenticateWithResponses performs the complete OAuth 2.0 PKCE authentication flow. It initiates authorization
// with PKCE parameters, submits credentials and handles optional TOTP verification, then exchanges
// the resulting authorization code for OIDC tokens. Additionally, it returns the *resty.Responses that
// completed during the OIDC flow. The values set are determined by how far the process is able to proceed.
func (e *EdgeOidcAuthenticator) AuthenticateWithResponses() (*oidc.Tokens[*oidc.IDTokenClaims], *OidcAuthResponses, error) {
	oidcAuthResponses := &OidcAuthResponses{}

	pkceParams, err := newPkceParameters()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate PKCE parameters: %w", err)
	}

	verificationParams, initResp, err := e.initOAuthFlow(pkceParams)

	oidcAuthResponses.InitResponse = initResp

	if err != nil {
		return nil, oidcAuthResponses, fmt.Errorf("failed to initiate authorization flow: %w", err)
	}

	authResponses, err := e.handlePrimaryAndSecondaryAuth(verificationParams)

	if authResponses != nil {
		oidcAuthResponses.PrimaryCredentialResponse = authResponses.PrimaryResp
		oidcAuthResponses.TotpEnrollResponse = authResponses.TotpEnrollResp
		oidcAuthResponses.TotpEnrollVerifyResponse = authResponses.TotpEnrollVerifyResp
		oidcAuthResponses.TotpResponse = authResponses.TotpResp
		oidcAuthResponses.RedirectResponse = authResponses.RedirectResp
	}
	if err != nil {
		return nil, oidcAuthResponses, err
	}

	tokens, err := e.finishOAuthFlow(oidcAuthResponses.RedirectResponse, verificationParams, pkceParams)
	if err != nil {
		return nil, oidcAuthResponses, err
	}

	return tokens, oidcAuthResponses, nil
}

// Authorize starts the OIDC PKCE authorization flow without completing it, returning an
// OidcAuthorizeResult for use when tests need to make raw HTTP calls to the OP login endpoints.
// The Exchange function in the result completes the flow by trading an authorization code for tokens.
func (e *EdgeOidcAuthenticator) Authorize() (*OidcAuthorizeResult, error) {
	pkceParams, err := newPkceParameters()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE parameters: %w", err)
	}

	verificationParams, _, err := e.initOAuthFlow(pkceParams)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate authorization flow: %w", err)
	}

	return &OidcAuthorizeResult{
		AuthRequestId: verificationParams.AuthRequestId,
		Client:        e.restyClient,
		Exchange: func(code string) (*oidc.Tokens[*oidc.IDTokenClaims], error) {
			return e.exchangeAuthorizationCodeForTokens(code, pkceParams)
		},
	}, nil
}

// finishOAuthFlow extracts the authorization code from the callback redirect and exchanges it for tokens.
// The authorization server returns the code as a query parameter in the Location header of the redirect response.
// The code is then used with the PKCE verifier to obtain OIDC tokens via the token endpoint.
func (e *EdgeOidcAuthenticator) finishOAuthFlow(redirectResp *resty.Response, verificationParams *verificationParameters, pkceParams *pkceParameters) (*oidc.Tokens[*oidc.IDTokenClaims], error) {
	if redirectResp.StatusCode() != http.StatusFound {
		return nil, fmt.Errorf("authentication failed, expected a 302, got %d", redirectResp.StatusCode())
	}

	redirectStr := redirectResp.Header().Get("Location")
	redirectUrl, err := url.Parse(redirectStr)
	if err != nil {
		return nil, fmt.Errorf("authentication failed, could not parse redirect url [%s]: %w", redirectStr, err)
	}

	state := redirectUrl.Query().Get("state")

	if state == "" {
		return nil, errors.New("authentication failed, no state found in redirect url")
	}

	if state != verificationParams.State {
		return nil, errors.New("authentication failed, state mismatch")
	}

	code := redirectUrl.Query().Get("code")
	if code == "" {
		return nil, errors.New("authentication failed, no code found in redirect url")
	}

	tokens, err := e.exchangeAuthorizationCodeForTokens(code, pkceParams)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	if tokens.IDTokenClaims.Nonce != verificationParams.Nonce {
		return nil, errors.New("authentication failed, nonce mismatch")
	}

	return tokens, nil
}

// PrimaryAndSecondaryAuthResponses holds the HTTP responses collected during
// the primary credential submission and optional secondary authentication
// steps of the OIDC flow. Fields are nil if their corresponding step was not reached.
type PrimaryAndSecondaryAuthResponses struct {
	RedirectResp         *resty.Response
	PrimaryResp          *resty.Response
	TotpEnrollResp       *resty.Response
	TotpEnrollVerifyResp *resty.Response
	TotpResp             *resty.Response
}

// handlePrimaryAndSecondaryAuth submits credentials to the authorization endpoint and handles optional TOTP.
func (e *EdgeOidcAuthenticator) handlePrimaryAndSecondaryAuth(verificationParams *verificationParameters) (*PrimaryAndSecondaryAuthResponses, error) {
	authResponses := &PrimaryAndSecondaryAuthResponses{}

	loginUri := "https://" + e.ApiHost + "/oidc/login/" + string(e.Credentials.Method())
	totpUri := "https://" + e.ApiHost + "/oidc/login/totp"

	payload := &authPayload{
		Authenticate:  e.Credentials.Payload(),
		AuthRequestId: verificationParams.AuthRequestId,
	}

	if e.ConfigTypeOverrides != nil {
		payload.ConfigTypes = e.ConfigTypeOverrides
	}

	formData := payload.toValues()
	req := e.restyClient.R()
	clientRequest := asClientRequest(req, e.restyClient)

	err := e.Credentials.AuthenticateRequest(clientRequest, strfmt.Default)
	if err != nil {
		return authResponses, err
	}

	resp, err := req.SetFormDataFromValues(formData).Post(loginUri)
	authResponses.PrimaryResp = resp
	if err != nil {
		return authResponses, err
	}

	// no additional secondary authentication required
	if resp.StatusCode() == http.StatusFound {
		authResponses.RedirectResp = resp
		return authResponses, nil
	}

	// something went wrong
	if resp.StatusCode() != http.StatusOK {
		return authResponses, fmt.Errorf("credential submission failed with status %d", resp.StatusCode())
	}

	totpRequiredHeader := resp.Header().Get(TotpRequiredHeader)
	if totpRequiredHeader == "" {
		return authResponses, errors.New("response was not a redirect and TOTP is not required, unknown additional authentication steps are required but unsupported")
	}

	// Parse the auth queries response body to determine if TOTP is already enrolled.
	var authQueriesResp struct {
		AuthQueries []*rest_model.AuthQueryDetail `json:"authQueries"`
	}
	isTotpEnrolled := true // default to enrolled; enrollment check requires a parseable body
	if err = json.Unmarshal(resp.Body(), &authQueriesResp); err == nil {
		for _, q := range authQueriesResp.AuthQueries {
			if q.TypeID == rest_model.AuthQueryTypeTOTP {
				isTotpEnrolled = q.IsTotpEnrolled
				break
			}
		}
	}

	if !isTotpEnrolled {
		return e.handleTotpEnrollment(authResponses, payload.AuthRequestId)
	}

	if e.TotpCodeProvider == nil {
		return authResponses, errors.New("totp is required but no totp callback was defined")
	}

	totpCodeResultCh := e.TotpCodeProvider.GetTotpCode()
	var totpCode string

	select {
	case totpCodeResult := <-totpCodeResultCh:
		if totpCodeResult.Err != nil {
			return nil, fmt.Errorf("error getting totp code: %w", totpCodeResult.Err)
		}
		totpCode = totpCodeResult.Code
	case <-time.After(30 * time.Minute):
		return nil, fmt.Errorf("timeout waiting for totp code provider")
	}

	resp, err = e.restyClient.R().SetBody(&totpCodePayload{
		MfaCode: rest_model.MfaCode{
			Code: &totpCode,
		},
		AuthRequestId: payload.AuthRequestId,
	}).Post(totpUri)

	authResponses.TotpResp = resp

	if err != nil {
		return authResponses, rest_util.WrapErr(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return authResponses, errors.New("totp code verified, but additional authentication is required that is not supported or not configured, cannot authenticate")
	case http.StatusFound:
		authResponses.RedirectResp = resp
		return authResponses, nil
	case http.StatusBadRequest:
		return authResponses, errors.New("totp code did not verify")
	default:
		return authResponses, fmt.Errorf("unexpected response code %d from TOTP verification", resp.StatusCode())
	}
}

// handleTotpEnrollment performs TOTP enrollment during an OIDC authentication flow.
// It starts enrollment to obtain a provisioning URL, delegates QR code display and code
// collection to the configured TotpEnrollmentProvider, then verifies the resulting code
// to complete enrollment.
func (e *EdgeOidcAuthenticator) handleTotpEnrollment(authResponses *PrimaryAndSecondaryAuthResponses, authRequestId string) (*PrimaryAndSecondaryAuthResponses, error) {
	if e.TotpEnrollmentProvider == nil {
		return authResponses, errors.New("totp enrollment is required but no totp enrollment provider was configured")
	}

	enrollUri := "https://" + e.ApiHost + "/oidc/login/totp/enroll"
	enrollVerifyUri := "https://" + e.ApiHost + "/oidc/login/totp/enroll/verify"

	enrollResp, err := e.restyClient.R().SetBody(&authRequestIdPayload{
		AuthRequestId: authRequestId,
	}).Post(enrollUri)

	authResponses.TotpEnrollResp = enrollResp

	if err != nil {
		return authResponses, fmt.Errorf("totp enrollment start failed: %w", err)
	}

	if enrollResp.StatusCode() != http.StatusCreated {
		return authResponses, fmt.Errorf("totp enrollment start failed with status %d", enrollResp.StatusCode())
	}

	var enrollDetail rest_model.DetailMfa
	if err = json.Unmarshal(enrollResp.Body(), &enrollDetail); err != nil {
		return authResponses, fmt.Errorf("failed to parse totp enrollment response: %w", err)
	}

	provisioningUrl := enrollDetail.ProvisioningURL
	if provisioningUrl == "" {
		return authResponses, errors.New("totp enrollment response did not contain a provisioning URL")
	}

	enrollResultCh := e.TotpEnrollmentProvider.GetTotpEnrollmentCode(provisioningUrl)
	var enrollCode string

	select {
	case enrollResult := <-enrollResultCh:
		if enrollResult.Err != nil {
			return authResponses, fmt.Errorf("totp enrollment cancelled: %w", enrollResult.Err)
		}
		enrollCode = enrollResult.Code
	case <-time.After(30 * time.Minute):
		return authResponses, errors.New("timeout waiting for totp enrollment code")
	}

	enrollVerifyResp, err := e.restyClient.R().SetBody(&totpCodePayload{
		MfaCode: rest_model.MfaCode{
			Code: &enrollCode,
		},
		AuthRequestId: authRequestId,
	}).Post(enrollVerifyUri)

	authResponses.TotpEnrollVerifyResp = enrollVerifyResp

	if err != nil {
		return authResponses, fmt.Errorf("totp enrollment verification failed: %w", err)
	}

	switch enrollVerifyResp.StatusCode() {
	case http.StatusFound:
		authResponses.RedirectResp = enrollVerifyResp
		return authResponses, nil
	case http.StatusOK:
		return authResponses, errors.New("totp enrollment verified, but additional authentication is required that is not supported or not configured, cannot authenticate")
	case http.StatusBadRequest:
		return authResponses, errors.New("totp enrollment code did not verify")
	default:
		return authResponses, fmt.Errorf("unexpected response code %d from totp enrollment verification", enrollVerifyResp.StatusCode())
	}
}

// initOAuthFlow initiates the OAuth authorization request with PKCE parameters and returns the authorization request ID.
func (e *EdgeOidcAuthenticator) initOAuthFlow(pkceParams *pkceParameters) (*verificationParameters, *resty.Response, error) {
	verificationParams := &verificationParameters{
		State: generateRandomState(),
		Nonce: generateNonce(),
	}

	authUrl := "https://" + e.ApiHost + "/oidc/authorize?" + url.Values{
		"client_id":             []string{"native"},
		"response_type":         []string{"code"},
		"scope":                 []string{"openid offline_access"},
		"state":                 []string{verificationParams.State},
		"code_challenge":        []string{pkceParams.Challenge},
		"code_challenge_method": []string{pkceParams.Method},
		"redirect_uri":          []string{e.RedirectUri},
		"nonce":                 []string{verificationParams.Nonce},
	}.Encode()

	resp, err := e.restyClient.R().Get(authUrl)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode() != http.StatusOK {
		body := resp.Body()

		if len(body) == 0 {
			body = []byte("<body was empty>")
		}

		return nil, resp, fmt.Errorf("authentication request start failed with status %d, either a misconfigured request was sent or the expected redirect URL (%s) is not allowed: %s", resp.StatusCode(), e.RedirectUri, body)
	}

	verificationParams.AuthRequestId = resp.Header().Get(AuthRequestIdHeader)
	if verificationParams.AuthRequestId == "" {
		return nil, resp, errors.New("could not find auth request id header from authorize endpoint")
	}

	return verificationParams, resp, nil
}

// RedirectUntilUrlPrefix returns a redirect policy that follows redirects until the request URL
// matches one of the provided URL prefixes. Once a matching prefix is encountered, the redirect
// is not followed, allowing the caller to inspect the redirect response.
func RedirectUntilUrlPrefix(urlPrefixToStopAt ...string) resty.RedirectPolicy {
	return resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		reqUrl := req.URL.String()
		for _, urlToStopAt := range urlPrefixToStopAt {
			if strings.HasPrefix(reqUrl, urlToStopAt) {
				return http.ErrUseLastResponse
			}
		}
		return nil
	})
}

// exchangeAuthorizationCodeForTokens exchanges an authorization code and PKCE verifier for OIDC tokens.
func (e *EdgeOidcAuthenticator) exchangeAuthorizationCodeForTokens(code string, pkceParams *pkceParameters) (*oidc.Tokens[*oidc.IDTokenClaims], error) {
	tokenEndpoint := "https://" + e.ApiHost + "/oidc/oauth/token"

	tokenResp, err := e.restyClient.R().SetFormData(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     "native",
		"code_verifier": pkceParams.Verifier,
		"code":          code,
		"redirect_uri":  DefaultOidcRedirectUri,
	}).Post(tokenEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code for tokens: %w", err)
	}

	if tokenResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", tokenResp.StatusCode(), string(tokenResp.Body()))
	}

	// Parse token response
	var tokenData map[string]interface{}
	err = json.Unmarshal(tokenResp.Body(), &tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	accessToken, ok := tokenData["access_token"].(string)
	if !ok {
		return nil, errors.New("access_token not found in token response")
	}

	refreshToken, _ := tokenData["refresh_token"].(string)
	expiresIn, _ := tokenData["expires_in"].(float64)

	// Parse ID token
	idToken, _ := tokenData["id_token"].(string)
	idClaims := &IdClaims{}

	if idToken != "" {
		_, _, err = jwt.NewParser().ParseUnverified(idToken, idClaims)
		if err != nil {
			// Log but don't fail if ID token parsing fails
			return nil, fmt.Errorf("failed to parse ID token: %w", err)
		}
	}

	tokens := &oidc.Tokens[*oidc.IDTokenClaims]{
		Token: &oauth2.Token{
			AccessToken:  accessToken,
			TokenType:    "Bearer",
			RefreshToken: refreshToken,
			Expiry:       time.Now().Add(time.Duration(expiresIn) * time.Second),
		},
		IDTokenClaims: &idClaims.IDTokenClaims,
		IDToken:       idToken,
	}

	return tokens, nil
}

// pkceParameters holds the PKCE parameters used for OAuth 2.0 Proof Key for Public Clients flow.
type pkceParameters struct {
	Verifier  string
	Challenge string
	Method    string
}

type verificationParameters struct {
	State         string
	AuthRequestId string
	Nonce         string
}

// newPkceParameters generates PKCE parameters for OAuth 2.0 PKCE flow.
// It creates a random code verifier and derives the code challenge by applying SHA256 hashing.
func newPkceParameters() (*pkceParameters, error) {
	var err error
	params := &pkceParameters{
		Method: "S256",
	}

	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	params.Verifier = base64URLEncodeNoPadding(b)

	hash := sha256.Sum256([]byte(params.Verifier))
	params.Challenge = base64URLEncodeNoPadding(hash[:])

	return params, nil
}

// generateRandomState generates a random state string for CSRF protection.
func generateRandomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64URLEncodeNoPadding(b)
}

// generateNonce generates a random nonce for binding the authorization request to the ID token.
func generateNonce() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// base64URLEncodeNoPadding encodes data to base64URL format without padding.
// Padding is removed because base64URL is designed to work in URLs and query strings where
// the '=' character may have special meaning.
func base64URLEncodeNoPadding(data []byte) string {
	encoded := base64.URLEncoding.EncodeToString(data)
	return strings.TrimRight(encoded, "=")
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
