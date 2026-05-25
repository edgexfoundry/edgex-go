package edge_apis

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/foundation/v2/stringz"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
)

var _ json.Marshaler = (*ApiSessionJsonWrapper)(nil)
var _ json.Unmarshaler = (*ApiSessionJsonWrapper)(nil)

// ApiSessionJsonWrapper provides JSON marshaling and unmarshaling capabilities for ApiSession
// interface types. It allows polymorphic ApiSession implementations (ApiSessionLegacy and
// ApiSessionOidc) to be correctly serialized and deserialized by delegating to the underlying
// ApiSession's JSON methods.
//
// This wrapper enables ApiSession instances to be embedded in structs and marshaled to/from
// JSON.
type ApiSessionJsonWrapper struct {
	ApiSession ApiSession
}

func (a *ApiSessionJsonWrapper) UnmarshalJSON(bytes []byte) error {
	var err error
	a.ApiSession, err = UnmarshalApiSession(bytes)

	return err
}

func (a *ApiSessionJsonWrapper) MarshalJSON() ([]byte, error) {
	return a.ApiSession.MarshalJSON()
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

	// GetType returns the authentication method used to establish this session, enabling
	// callers to determine whether legacy or OIDC-based authentication is in use.
	GetType() ApiSessionType

	json.Marshaler
	json.Unmarshaler
}

type ApiSessionJson struct {
	Type             string `json:"type"`
	ZtSessionToken   string `json:"ztSessionToken,omitempty"`
	OidcAccessToken  string `json:"oidcAccessToken,omitempty"`
	OidcRefreshToken string `json:"oidcRefreshToken,omitempty"`
}

// ApiSessionType identifies the authentication mechanism used to establish an API session.
type ApiSessionType string

const (
	// ApiSessionTypeLegacy indicates a session created using the original Ziti authentication
	// with session tokens passed in the zt-session header.
	ApiSessionTypeLegacy ApiSessionType = "legacy"

	// ApiSessionTypeOidc indicates a session created using OpenID Connect authentication
	// with JWT bearer tokens.
	ApiSessionTypeOidc ApiSessionType = "oidc"
)

func UnmarshalApiSession(data []byte) (ApiSession, error) {
	apiSessionJson := &ApiSessionJson{}

	err := json.Unmarshal(data, apiSessionJson)

	if err != nil {
		return nil, err
	}

	switch apiSessionJson.Type {
	case string(ApiSessionTypeLegacy):
		result := &ApiSessionLegacy{}
		err := result.setFromJson(apiSessionJson)
		if err != nil {
			return nil, err
		}
		return result, nil
	case string(ApiSessionTypeOidc):
		result := &ApiSessionOidc{}
		err := result.setFromJson(apiSessionJson)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	return nil, fmt.Errorf("unsupported api session type %s", apiSessionJson.Type)
}

var _ ApiSession = (*ApiSessionLegacy)(nil)
var _ ApiSession = (*ApiSessionOidc)(nil)

// ApiSessionLegacy represents OpenZiti's original authentication API Session Detail, supplied in the `zt-session` header.
// It has been supplanted by OIDC authentication represented by ApiSessionOidc.
type ApiSessionLegacy struct {
	Detail         *rest_model.CurrentAPISessionDetail
	RequestHeaders http.Header
}

func NewApiSessionLegacy(token string) *ApiSessionLegacy {
	return &ApiSessionLegacy{
		Detail: &rest_model.CurrentAPISessionDetail{
			APISessionDetail: rest_model.APISessionDetail{
				Token: &token,
			},
		},
	}
}

func (a *ApiSessionLegacy) NewApiSessionLegacy(token string) *ApiSessionLegacy {
	return &ApiSessionLegacy{
		Detail: &rest_model.CurrentAPISessionDetail{
			APISessionDetail: rest_model.APISessionDetail{
				Token: &token,
			},
		},
	}
}

func (a *ApiSessionLegacy) GetType() ApiSessionType {
	return ApiSessionTypeLegacy
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

func (a *ApiSessionLegacy) MarshalJSON() ([]byte, error) {
	apiSessionJson := ApiSessionJson{
		Type:           string(a.GetType()),
		ZtSessionToken: string(a.GetToken()),
	}

	return json.Marshal(apiSessionJson)
}

func (a *ApiSessionLegacy) UnmarshalJSON(bytes []byte) error {
	apiSessionJson := ApiSessionJson{}
	err := json.Unmarshal(bytes, &apiSessionJson)
	if err != nil {
		return err
	}

	return a.setFromJson(&apiSessionJson)
}

func (a *ApiSessionLegacy) setFromJson(apiSessionJson *ApiSessionJson) error {
	if apiSessionJson.Type != string(ApiSessionTypeLegacy) {
		return fmt.Errorf("unsupported api session type %s", apiSessionJson.Type)
	}

	a.Detail = &rest_model.CurrentAPISessionDetail{
		APISessionDetail: rest_model.APISessionDetail{
			Token: &apiSessionJson.ZtSessionToken,
		},
	}

	return nil
}

// ApiSessionOidc represents an authenticated session backed by OIDC tokens.
type ApiSessionOidc struct {
	OidcTokens     *oidc.Tokens[*oidc.IDTokenClaims]
	RequestHeaders http.Header
}

func NewApiSessionOidc(accessToken, refreshToken string) *ApiSessionOidc {
	return &ApiSessionOidc{
		OidcTokens: &oidc.Tokens[*oidc.IDTokenClaims]{
			Token: &oauth2.Token{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		},
	}
}

func (a *ApiSessionOidc) GetType() ApiSessionType {
	return ApiSessionTypeOidc
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

func (a *ApiSessionOidc) MarshalJSON() ([]byte, error) {
	apiSessionJson := &ApiSessionJson{
		Type:             string(a.GetType()),
		OidcAccessToken:  a.OidcTokens.AccessToken,
		OidcRefreshToken: a.OidcTokens.RefreshToken,
	}

	return json.Marshal(apiSessionJson)
}

func (a *ApiSessionOidc) UnmarshalJSON(bytes []byte) error {
	apiSessionJson := &ApiSessionJson{}

	err := json.Unmarshal(bytes, &apiSessionJson)
	if err != nil {
		return err
	}

	if apiSessionJson.Type != string(ApiSessionTypeOidc) {
		return fmt.Errorf("unsupported api session type %s", apiSessionJson.Type)
	}

	return a.setFromJson(apiSessionJson)
}

func (a *ApiSessionOidc) setFromJson(apiSessionJson *ApiSessionJson) error {
	a.OidcTokens = &oidc.Tokens[*oidc.IDTokenClaims]{
		Token: &oauth2.Token{
			AccessToken:  apiSessionJson.OidcAccessToken,
			RefreshToken: apiSessionJson.OidcRefreshToken,
		},
	}

	return nil
}
