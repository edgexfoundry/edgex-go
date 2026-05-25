package edge_apis

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// JwtTokenPrefix is the standard prefix for JWT tokens, representing the first two characters
// of a Base64URL-encoded JWT header. This prefix is used to identify JWT-format tokens.
const JwtTokenPrefix = "ey"

// ServiceAccessClaims represents the JWT claims for service-level access tokens, including
// identity and session binding information specific to a service connection.
type ServiceAccessClaims struct {
	jwt.RegisteredClaims
	ApiSessionId string `json:"z_asid"`
	IdentityId   string `json:"z_iid"`
	TokenType    string `json:"z_t"`
	Type         string `json:"z_st"`
}

// ApiAccessClaims represents the JWT claims for API session access tokens, including
// identity attributes, administrative status, and configuration bindings.
type ApiAccessClaims struct {
	jwt.RegisteredClaims
	ApiSessionId     string   `json:"z_asid,omitempty"`
	ExternalId       string   `json:"z_eid,omitempty"`
	IsAdmin          bool     `json:"z_ia,omitempty"`
	ConfigTypes      []string `json:"z_ct,omitempty"`
	ApplicationId    string   `json:"z_aid,omitempty"`
	Type             string   `json:"z_t"`
	CertFingerprints []string `json:"z_cfs"`
	Scopes           []string `json:"scopes,omitempty"`
}

var _ jwt.Claims = (*IdClaims)(nil)

// IdClaims wraps oidc.IDToken claims to fulfill the jwt.Claims interface
type IdClaims struct {
	oidc.IDTokenClaims
}

func (r *IdClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: r.GetExpiration()}, nil
}

func (r *IdClaims) GetNotBefore() (*jwt.NumericDate, error) {
	notBefore := r.NotBefore.AsTime()
	return &jwt.NumericDate{Time: notBefore}, nil
}

func (r *IdClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: r.TokenClaims.GetIssuedAt()}, nil
}

func (r *IdClaims) GetIssuer() (string, error) {
	return r.Issuer, nil
}

func (r *IdClaims) GetSubject() (string, error) {
	return r.Issuer, nil
}

func (r *IdClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings(r.Audience), nil
}
