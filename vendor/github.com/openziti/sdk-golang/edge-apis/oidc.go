package edge_apis

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/michaelquigley/pfxlog"
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const JwtTokenPrefix = "ey"

type ServiceAccessClaims struct {
	jwt.RegisteredClaims
	ApiSessionId string `json:"z_asid"`
	IdentityId   string `json:"z_iid"`
	TokenType    string `json:"z_t"`
	Type         string `json:"z_st"`
}

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
	return &jwt.NumericDate{Time: r.TokenClaims.GetExpiration()}, nil
}

func (r *IdClaims) GetNotBefore() (*jwt.NumericDate, error) {
	notBefore := r.TokenClaims.NotBefore.AsTime()
	return &jwt.NumericDate{Time: notBefore}, nil
}

func (r *IdClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: r.TokenClaims.GetIssuedAt()}, nil
}

func (r *IdClaims) GetIssuer() (string, error) {
	return r.TokenClaims.Issuer, nil
}

func (r *IdClaims) GetSubject() (string, error) {
	return r.TokenClaims.Issuer, nil
}

func (r *IdClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings(r.TokenClaims.Audience), nil
}

type localRpServer struct {
	Server       *http.Server
	Port         string
	Listener     net.Listener
	TokenChan    chan *oidc.Tokens[*oidc.IDTokenClaims]
	CallbackPath string
	CallbackUri  string
	LoginUri     string
}

func (t *localRpServer) Stop() {
	_ = t.Server.Shutdown(context.Background())
	close(t.TokenChan)
}

func (t *localRpServer) Start() {
	go func() {
		_ = t.Server.Serve(t.Listener)
	}()

	started := make(chan struct{})

	go func() {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		end := time.Now().Add(11 * time.Second)
		for {
			if time.Now().After(end) {
				break
			}

			time.Sleep(100 * time.Millisecond)

			_, err := client.Get(t.LoginUri)

			if err == nil {
				break
			}
		}
		close(started)
	}()
	select {
	case <-started:
	case <-time.After(10 * time.Second):
		pfxlog.Logger().Warn("local relying party server did not start within 10s")
	}
}

func newLocalRpServer(apiHost string, authMethod string) (*localRpServer, error) {
	tokenOutChan := make(chan *oidc.Tokens[*oidc.IDTokenClaims], 1)
	result := &localRpServer{
		CallbackPath: "/auth/callback",
		TokenChan:    tokenOutChan,
	}
	var err error

	result.Listener, err = net.Listen("tcp", ":0")

	if err != nil {
		return nil, fmt.Errorf("could not listen on a random port: %w", err)
	}

	_, result.Port, _ = net.SplitHostPort(result.Listener.Addr().String())

	result.LoginUri = "http://127.0.0.1:" + result.Port + "/login"

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("could not generate secure cookie key: %w", err)
	}

	urlBase := "https://" + apiHost
	issuer := urlBase + "/oidc"
	clientID := "native"
	clientSecret := ""
	scopes := []string{"openid", "offline_access"}
	result.CallbackUri = "http://127.0.0.1:" + result.Port + result.CallbackPath

	cookieHandler := httphelper.NewCookieHandler(key, key, httphelper.WithUnsecure())
	jar, _ := cookiejar.New(&cookiejar.Options{})
	httpClient := &http.Client{

		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: nil,
		Jar:           jar,
		Timeout:       10 * time.Second,
	}

	options := []rp.Option{
		rp.WithHTTPClient(httpClient),
		rp.WithPKCE(cookieHandler),
	}

	provider, err := rp.NewRelyingPartyOIDC(issuer, clientID, clientSecret, result.CallbackUri, scopes, options...)

	if err != nil {
		return nil, fmt.Errorf("could not create rp OIDC: %w", err)
	}

	state := func() string {
		return uuid.New().String()
	}
	serverMux := http.NewServeMux()

	authHandler := rp.AuthURLHandler(state, provider, rp.WithPromptURLParam("Welcome back!"), rp.WithURLParam("method", authMethod))
	loginHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authHandler.ServeHTTP(writer, request)
	})

	serverMux.Handle("/login", loginHandler)

	marshalToken := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, relyingParty rp.RelyingParty) {
		tokenOutChan <- tokens
		_, _ = w.Write([]byte("done!"))
	}

	serverMux.Handle(result.CallbackPath, rp.CodeExchangeHandler(marshalToken, provider))

	result.Server = &http.Server{Handler: serverMux}

	return result, nil
}
