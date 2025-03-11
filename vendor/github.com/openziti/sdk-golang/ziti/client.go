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

package ziti

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/edge-api/rest_client_api_client/authentication"
	"github.com/openziti/edge-api/rest_client_api_client/current_api_session"
	"github.com/openziti/edge-api/rest_client_api_client/current_identity"
	"github.com/openziti/edge-api/rest_client_api_client/informational"
	"github.com/openziti/edge-api/rest_client_api_client/posture_checks"
	"github.com/openziti/edge-api/rest_client_api_client/service"
	"github.com/openziti/edge-api/rest_client_api_client/session"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/edge-api/rest_util"
	"github.com/openziti/foundation/v2/genext"
	nfPem "github.com/openziti/foundation/v2/pem"
	"github.com/openziti/foundation/v2/versions"
	"github.com/openziti/identity"
	apis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/sdk-golang/ziti/edge/posture"
	"github.com/openziti/transport/v2"
	"github.com/pkg/errors"
	"strings"
	"sync/atomic"
)

// CtrlClient is a stateful version of ZitiEdgeClient that simplifies operations
type CtrlClient struct {
	*apis.ClientApiClient
	Credentials apis.Credentials

	lastServiceUpdate *strfmt.DateTime

	ApiSessionCertificateDetail rest_model.CurrentAPISessionCertificateDetail
	ApiSessionCsr               x509.CertificateRequest
	ApiSessionCertificate       *x509.Certificate
	ApiSessionPrivateKey        *ecdsa.PrivateKey
	ApiSessionCertInstance      string

	PostureCache                     *posture.Cache
	ConfigTypes                      []string
	supportsConfigTypesOnServiceList atomic.Bool
	capabilitiesLoaded               atomic.Bool
}

// GetCurrentApiSession returns the current cached ApiSession or nil
func (self *CtrlClient) GetCurrentApiSession() apis.ApiSession {
	return self.ClientApiClient.GetCurrentApiSession()
}

// Refresh will contact the controller extending the current ApiSession for legacy API Sessions
func (self *CtrlClient) Refresh() (apis.ApiSession, error) {
	if apiSession := self.GetCurrentApiSession(); apiSession != nil {
		newApiSession, err := self.API.RefreshApiSession(apiSession, self.HttpClient)

		if err != nil {
			return nil, err
		}

		self.ApiSession.Store(&newApiSession)

		return newApiSession, nil
	}

	return nil, errors.New("no api session")
}

// IsServiceListUpdateAvailable will contact the controller to determine if a new set of services are available. Service
// updates could entail gaining/losing services access via policy or runtime authorization revocation due to posture
// checks.
func (self *CtrlClient) IsServiceListUpdateAvailable() (bool, *strfmt.DateTime, error) {
	resp, err := self.API.CurrentAPISession.ListServiceUpdates(current_api_session.NewListServiceUpdatesParams(), self.GetCurrentApiSession())

	if err != nil {
		return true, nil, err
	}

	return self.lastServiceUpdate == nil || !resp.Payload.Data.LastChangeAt.Equal(*self.lastServiceUpdate), resp.Payload.Data.LastChangeAt, nil
}

// Authenticate attempts to use authenticate, overwriting any existing ApiSession.
func (self *CtrlClient) Authenticate() (apis.ApiSession, error) {
	var err error

	self.ApiSessionCertificate = nil

	apiSession, err := self.ClientApiClient.Authenticate(self.Credentials, self.ConfigTypes)

	if err != nil {
		return nil, rest_util.WrapErr(err)
	}

	_, err = self.GetIdentity()
	if err != nil {
		return nil, rest_util.WrapErr(err)
	}

	return apiSession, nil
}

// AuthenticateMFA handles MFA authentication queries may be provided. AuthenticateMFA allows
// the current identity for their current api session to attempt to pass MFA authentication.
func (self *CtrlClient) AuthenticateMFA(code string) error {
	params := authentication.NewAuthenticateMfaParams()
	params.MfaAuth = &rest_model.MfaCode{
		Code: &code,
	}
	_, err := self.API.Authentication.AuthenticateMfa(params, self.GetCurrentApiSession())

	if err != nil {
		return rest_util.WrapErr(err)
	}

	return nil
}

// SendPostureResponse creates a posture response (some state data the controller has requested) for services. This
// information is used to determine runtime authorization access to services via posture checks.
func (self *CtrlClient) SendPostureResponse(response rest_model.PostureResponseCreate) error {
	params := posture_checks.NewCreatePostureResponseParams()
	params.PostureResponse = response
	_, err := self.API.PostureChecks.CreatePostureResponse(params, self.GetCurrentApiSession())

	if err != nil {
		return rest_util.WrapErr(err)
	}
	return nil
}

// SendPostureResponseBulk provides the same functionality as SendPostureResponse but allows multiple responses
// to be sent in a single request.
func (self *CtrlClient) SendPostureResponseBulk(responses []rest_model.PostureResponseCreate) error {
	params := posture_checks.NewCreatePostureResponseBulkParams()
	params.PostureResponse = responses
	_, err := self.API.PostureChecks.CreatePostureResponseBulk(params, self.GetCurrentApiSession())

	if err != nil {
		return rest_util.WrapErr(err)
	}
	return nil
}

// GetCurrentIdentity returns the rest_model.IdentityDetail for the currently authenticated ApiSession.
func (self *CtrlClient) GetCurrentIdentity() (*rest_model.IdentityDetail, error) {
	params := current_identity.NewGetCurrentIdentityParams()
	resp, err := self.API.CurrentIdentity.GetCurrentIdentity(params, self.GetCurrentApiSession())

	if err != nil {
		return nil, rest_util.WrapErr(err)
	}

	return resp.Payload.Data, nil
}

// GetSession returns the full rest_model.SessionDetail for a specific id. Does not function with JWT backed sessions.
func (self *CtrlClient) GetSession(id string) (*rest_model.SessionDetail, error) {
	params := session.NewDetailSessionParams()
	params.ID = id
	resp, err := self.API.Session.DetailSession(params, self.GetCurrentApiSession())

	if err != nil {
		return nil, rest_util.WrapErr(err)
	}

	self.sanitizeSessionUrls(resp.Payload.Data)
	return resp.Payload.Data, nil
}

func (self *CtrlClient) GetSessionFromJwt(sessionToken string) (*rest_model.SessionDetail, error) {
	parser := jwt.NewParser()
	serviceAccessClaims := &apis.ServiceAccessClaims{}

	_, _, err := parser.ParseUnverified(sessionToken, serviceAccessClaims)

	if err != nil {
		return nil, err
	}

	params := service.NewListServiceEdgeRoutersParams()
	params.SessionToken = &sessionToken
	params.ID = serviceAccessClaims.Subject //service id

	resp, err := self.API.Service.ListServiceEdgeRouters(params, self.GetCurrentApiSession())

	if err != nil {
		return nil, rest_util.WrapErr(err)
	}
	createdAt := strfmt.DateTime(serviceAccessClaims.IssuedAt.Time)
	sessionType := rest_model.DialBind(serviceAccessClaims.Type)

	sessionDetail := &rest_model.SessionDetail{
		BaseEntity: rest_model.BaseEntity{
			Links:     nil,
			CreatedAt: &createdAt,
			ID:        &serviceAccessClaims.ID,
		},
		APISessionID: &serviceAccessClaims.ApiSessionId,
		IdentityID:   &serviceAccessClaims.IdentityId,
		ServiceID:    &serviceAccessClaims.Subject,
		Token:        &sessionToken,
		Type:         &sessionType,
	}

	for _, er := range resp.Payload.Data.EdgeRouters {
		sessionDetail.EdgeRouters = append(sessionDetail.EdgeRouters, &rest_model.SessionEdgeRouter{
			CommonEdgeRouterProperties: *er,
		})
	}

	self.sanitizeSessionUrls(sessionDetail)

	return sessionDetail, nil
}

// GetIdentity returns the identity.Identity used to facilitate authentication. Each identity.Identity instance
// may provide authentication material in the form of x509 certificates and private keys and/or trusted CA pools.
func (self *CtrlClient) GetIdentity() (identity.Identity, error) {
	if idProvider, ok := self.Credentials.(apis.IdentityProvider); ok {
		return idProvider.GetIdentity(), nil
	}

	if self.ApiSessionCertificate == nil {
		err := self.EnsureApiSessionCertificate()

		if err != nil {
			return nil, fmt.Errorf("could not ensure an API Session certificate is available: %v", err)
		}
	}

	return identity.NewClientTokenIdentityWithPool([]*x509.Certificate{self.ApiSessionCertificate}, self.ApiSessionPrivateKey, self.HttpTransport.TLSClientConfig.RootCAs), nil
}

// EnsureApiSessionCertificate will create an ApiSessionCertificate if one does not already exist.
func (self *CtrlClient) EnsureApiSessionCertificate() error {
	if self.ApiSessionCertificate == nil {
		return self.NewApiSessionCertificate()
	}

	return nil
}

// NewApiSessionCertificate will create a new ephemeral private key used to generate an ephemeral certificate
// that may be used with the current ApiSession. The generated certificate and private key are scoped to the
// ApiSession used to create it.
func (self *CtrlClient) NewApiSessionCertificate() error {
	if self.ApiSessionCertInstance == "" {
		self.ApiSessionCertInstance = uuid.NewString()
	}

	if self.ApiSessionPrivateKey == nil {
		var err error
		self.ApiSessionPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

		if err != nil {
			return fmt.Errorf("could not generate private key for api session certificate: %v", err)
		}
	}

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			Organization:       []string{"Ziti SDK"},
			OrganizationalUnit: []string{"golang"},
			CommonName:         "golang-sdk-" + self.ApiSessionCertInstance + "-" + uuid.NewString(),
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, self.ApiSessionPrivateKey)
	if err != nil {
		panic(err)
	}
	block := &pem.Block{
		Type:    "CERTIFICATE REQUEST",
		Headers: nil,
		Bytes:   csrBytes,
	}
	csrPemString := string(pem.EncodeToMemory(block))

	params := current_api_session.NewCreateCurrentAPISessionCertificateParams()
	params.SessionCertificate = &rest_model.CurrentAPISessionCertificateCreate{
		Csr: &csrPemString,
	}

	resp, err := self.API.CurrentAPISession.CreateCurrentAPISessionCertificate(params, self.GetCurrentApiSession())

	if err != nil {
		return rest_util.WrapErr(err)
	}

	certs := nfPem.PemBytesToCertificates([]byte(*resp.Payload.Data.Certificate))

	if len(certs) == 0 {
		return fmt.Errorf("expected at least 1 certificate creating an API Session Certificate, got 0")
	}

	pfxlog.Logger().Infof("new API Session Certificate: %x", sha1.Sum(certs[0].Raw))

	self.ApiSessionCertificate = certs[0]

	return nil
}

// GetServices will fetch the list of services that the identity of the current ApiSession has access to for dialing
// or binding.
func (self *CtrlClient) GetServices() ([]*rest_model.ServiceDetail, error) {
	params := service.NewListServicesParams()

	pageOffset := int64(0)

	pageLimit := int64(500)
	params.Limit = &pageLimit

	if self.supportsSetOfConfigTypesOnServiceList() {
		params.ConfigTypes = self.ConfigTypes
	}

	var services []*rest_model.ServiceDetail

	for {
		params.Offset = &pageOffset
		resp, err := self.API.Service.ListServices(params, self.GetCurrentApiSession())

		if err != nil {
			return nil, rest_util.WrapErr(err)
		}

		if services == nil {
			services = make([]*rest_model.ServiceDetail, 0, *resp.Payload.Meta.Pagination.TotalCount)
		}

		services = append(services, resp.Payload.Data...)

		pageOffset += pageLimit
		if pageOffset >= *resp.Payload.Meta.Pagination.TotalCount {
			break
		}
	}

	return services, nil
}

// GetService will fetch the specific service requested. If the service doesn't exist,
// nil will be returned
func (self *CtrlClient) GetService(name string) (*rest_model.ServiceDetail, error) {
	params := service.NewListServicesParams()

	filter := fmt.Sprintf(`name="%s"`, name)
	params.Filter = &filter

	resp, err := self.API.Service.ListServices(params, nil)

	if err != nil {
		return nil, rest_util.WrapErr(err)
	}

	if len(resp.Payload.Data) > 0 {
		return resp.Payload.Data[0], nil
	}

	return nil, nil
}

// GetServiceTerminators returns the client terminator details for a specific service.
func (self *CtrlClient) GetServiceTerminators(svc *rest_model.ServiceDetail, offset int, limit int) ([]*rest_model.TerminatorClientDetail, int, error) {
	params := service.NewListServiceTerminatorsParams()

	pageOffset := int64(offset)
	params.Offset = &pageOffset

	pageLimit := int64(limit)
	params.Limit = &pageLimit

	params.ID = *svc.ID

	resp, err := self.API.Service.ListServiceTerminators(params, self.GetCurrentApiSession())

	if err != nil {
		return nil, 0, rest_util.WrapErr(err)
	}

	return resp.Payload.Data, int(*resp.Payload.Meta.Pagination.TotalCount), nil
}

// CreateSession will attempt to obtain a session token for a specific service id and type.
func (self *CtrlClient) CreateSession(id string, sessionType SessionType) (*rest_model.SessionDetail, error) {
	params := session.NewCreateSessionParams()
	params.Session = &rest_model.SessionCreate{
		ServiceID: id,
		Type:      rest_model.DialBind(sessionType),
	}

	resp, err := self.API.Session.CreateSession(params, self.GetCurrentApiSession())

	if err != nil {
		return nil, rest_util.WrapErr(err)
	}

	self.sanitizeSessionUrls(resp.Payload.Data)
	return resp.Payload.Data, nil
}

// EnrollMfa will attempt to start TOTP MFA enrollment for the currently authenticated identity.
func (self *CtrlClient) EnrollMfa() (*rest_model.DetailMfa, error) {
	enrollMfaParams := current_identity.NewEnrollMfaParams()

	apiSession := self.GetCurrentApiSession()

	_, enrollMfaErr := self.API.CurrentIdentity.EnrollMfa(enrollMfaParams, apiSession)

	if enrollMfaErr != nil {
		return nil, enrollMfaErr
	}

	detailMfaParams := current_identity.NewDetailMfaParams()
	detailMfaResp, detailMfaErr := self.API.CurrentIdentity.DetailMfa(detailMfaParams, apiSession)

	if detailMfaErr != nil {
		return nil, rest_util.WrapErr(detailMfaErr)
	}

	return detailMfaResp.Payload.Data, nil
}

// VerifyMfa will complete a TOTP MFA enrollment created via EnrollMfa.
func (self *CtrlClient) VerifyMfa(code string) error {
	params := current_identity.NewVerifyMfaParams()

	params.MfaValidation = &rest_model.MfaCode{
		Code: &code,
	}

	_, err := self.API.CurrentIdentity.VerifyMfa(params, self.GetCurrentApiSession())

	return rest_util.WrapErr(err)
}

// RemoveMfa will remove the currently enrolled TOTP MFA added by EnrollMfa() and verified by VerifyMfa()
func (self *CtrlClient) RemoveMfa(code string) error {
	params := current_identity.NewDeleteMfaParams()
	params.MfaValidationCode = &code

	_, err := self.API.CurrentIdentity.DeleteMfa(params, self.GetCurrentApiSession())

	return rest_util.WrapErr(err)
}

// sanitizeSessionUrls will transform ER urls to transport friendly URIs and remove
// any addresses that cannot be parsed
func (self *CtrlClient) sanitizeSessionUrls(session *rest_model.SessionDetail) {
	for _, edgeRouter := range session.EdgeRouters {
		newUrls := map[string]string{}
		for protocol, url := range edgeRouter.SupportedProtocols {
			url = strings.Replace(url, "://", ":", 1)
			if _, err := transport.ParseAddress(url); err == nil {
				newUrls[protocol] = url
			} else {
				pfxlog.Logger().WithError(err).Debugf("ignoring address [%s] for router [%s], as it can't be parsed", url, genext.OrDefault(edgeRouter.Name))
			}
		}
		edgeRouter.SupportedProtocols = newUrls
	}
}

func (self *CtrlClient) loadCtrlCapabilities() {
	result, _ := self.API.Informational.ListVersion(informational.NewListVersionParams())
	if result != nil && result.Payload != nil && result.Payload.Data != nil {
		if sv, err := versions.ParseSemVer(result.Payload.Data.Version); err == nil {
			if sv.Equals(versions.MustParseSemVer("0.0.0")) || sv.CompareTo(versions.MustParseSemVer("1.1.0")) >= 0 {
				self.supportsConfigTypesOnServiceList.Store(true)
			}
		}
	}
	self.capabilitiesLoaded.Store(true)
}

func (self *CtrlClient) supportsSetOfConfigTypesOnServiceList() bool {
	if !self.capabilitiesLoaded.Load() {
		self.loadCtrlCapabilities()
	}
	return self.supportsConfigTypesOnServiceList.Load()
}
