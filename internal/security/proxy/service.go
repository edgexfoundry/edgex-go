/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @author: Tingyu Zeng, Dell
 *******************************************************************************/
package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type CertUploadErrorType int

const (
	CertExisting  CertUploadErrorType = 0
	InternalError CertUploadErrorType = 1
)

type CertError struct {
	err    string
	reason CertUploadErrorType
}

func (ce *CertError) Error() string {
	return ce.err
}

type Service struct {
	client        internal.HttpCaller
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct
}

func NewService(
	r internal.HttpCaller,
	loggingClient logger.LoggingClient,
	configuration *config.ConfigurationStruct) Service {

	return Service{
		client:        r,
		loggingClient: loggingClient,
		configuration: configuration,
	}
}

func (s *Service) CheckProxyServiceStatus() error {
	return s.checkServiceStatus(s.configuration.KongURL.GetProxyBaseURL())
}

func (s *Service) CheckSecretServiceStatus() error {
	return s.checkServiceStatus(s.configuration.SecretService.GetSecretSvcBaseURL())
}

func (s *Service) checkServiceStatus(path string) error {
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("the status of service on %s is unknown, the initialization is terminated", path)
		return errors.New(e)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		s.loggingClient.Info(fmt.Sprintf("the service on %s is up successfully", path))
		break
	default:
		err = fmt.Errorf("unexpected http status %v %s", resp.StatusCode, path)
		s.loggingClient.Error(err.Error())
		return err
	}
	return nil
}

func (s *Service) ResetProxy() error {
	paths := []string{RoutesPath, ServicesPath, ConsumersPath, PluginsPath, CertificatesPath}
	for _, path := range paths {
		d, err := s.getSvcIDs(path)
		if err != nil {
			return err
		}
		for _, c := range d.Section {
			r := NewResource(c.ID, s.client, s.configuration.KongURL.GetProxyBaseURL(), s.loggingClient)
			err = r.Remove(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Init(cert CertificateLoader) error {
	postErr := s.postCert(cert)
	if postErr != nil {
		switch postErr.reason {
		case CertExisting:
			s.loggingClient.Info("skipping as the initialization has been done successfully")
			return nil
		case InternalError:
			return errors.New(postErr.Error())
		default:
			return errors.New(postErr.Error())
		}
	}

	for clientName, client := range s.configuration.Clients {
		serviceParams := &KongService{
			Name:     strings.ToLower(clientName),
			Host:     client.Host,
			Port:     client.Port,
			Protocol: client.Protocol,
		}

		err := s.initKongService(serviceParams)
		if err != nil {
			return err
		}

		routeParams := &KongRoute{
			Paths: []string{"/" + strings.ToLower(clientName)},
			Name:  strings.ToLower(clientName),
		}
		err = s.initKongRoutes(routeParams, strings.ToLower(clientName))
		if err != nil {
			return err
		}
	}

	err := s.initAuthMethod(s.configuration.KongAuth.Name, s.configuration.KongAuth.TokenTTL)
	if err != nil {
		return err
	}

	err = s.initACL(s.configuration.KongACL.Name, s.configuration.KongACL.WhiteList)
	if err != nil {
		return err
	}

	s.loggingClient.Info("finishing initialization for reverse proxy")
	return nil
}

func (s *Service) postCert(cert CertificateLoader) *CertError {
	cp, err := cert.Load()

	if err != nil {
		return &CertError{err.Error(), InternalError}
	}

	body := &CertInfo{
		Cert: cp.Cert,
		Key:  cp.Key,
		Snis: s.configuration.SecretService.SNIS,
	}
	s.loggingClient.Debug("trying to upload cert to proxy server")
	data, err := json.Marshal(body)
	if err != nil {
		s.loggingClient.Error(err.Error())
		return &CertError{err.Error(), InternalError}
	}
	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), CertificatesPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(string(data)))
	req.Header.Add(clients.ContentType, clients.ContentTypeJSON)
	if err != nil {
		s.loggingClient.Error("failed to create upload cert request -- %s", err.Error())
		return &CertError{err.Error(), InternalError}
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.loggingClient.Error("failed to upload cert to proxy server with error %s", err.Error())
		return &CertError{err.Error(), InternalError}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		s.loggingClient.Info("successfully added certificate to the reverse proxy")
		break
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &CertError{err.Error(), InternalError}
		}
		why := string(b)
		message := fmt.Sprintf("failed to add certificate with errorcode %d, error %s", resp.StatusCode, why)
		s.loggingClient.Error(message)
		e := &CertError{message, InternalError}
		if (resp.StatusCode == http.StatusBadRequest) && (strings.Index(why, "existing certificate") != -1) {
			message = fmt.Sprintf("certificate already exists on reverse proxy")
		}
		e = &CertError{message, CertExisting}
		return e
	}
	return nil
}

func (s *Service) initKongService(service *KongService) error {
	formVals := url.Values{
		"name":     {service.Name},
		"host":     {service.Host},
		"port":     {strconv.Itoa(service.Port)},
		"protocol": {service.Protocol},
	}
	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), ServicesPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		return fmt.Errorf("failed to construct http POST form request: %s %s", service.Name, err.Error())
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up proxy service: %s %s", service.Name, err.Error())
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		s.loggingClient.Info(fmt.Sprintf("successful to set up proxy service for %s", service.Name))
		break
	case http.StatusConflict:
		s.loggingClient.Info(fmt.Sprintf("proxy service for %s has been set up", service.Name))
		break
	default:
		err = fmt.Errorf("proxy service for %s returned status %d", service.Name, resp.StatusCode)
		s.loggingClient.Error(err.Error())
		return err
	}
	return nil
}

func (s *Service) initKongRoutes(r *KongRoute, name string) error {
	data, err := json.Marshal(r)
	if err != nil {
		s.loggingClient.Error(err.Error())
		return err
	}
	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), ServicesPath, name, "routes"}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(string(data)))
	if err != nil {
		e := fmt.Sprintf("failed to set up routes for %s with error %s", name, err.Error())
		s.loggingClient.Error(e)
		return err
	}
	req.Header.Add(clients.ContentType, clients.ContentTypeJSON)

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up routes for %s with error %s", name, err.Error())
		s.loggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		s.loggingClient.Info(fmt.Sprintf("successful to set up route for %s", name))
		break
	default:
		e := fmt.Sprintf("failed to set up route for %s with error %s", name, resp.Status)
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) initACL(name string, whitelist string) error {
	aclParams := &KongACLPlugin{
		Name:      name,
		WhiteList: whitelist,
	}
	// The type above is largely useless but I'm leaving it for now as otherwise there'd be no way to know
	// where or why the second field below is "config.whitelist"
	formVals := url.Values{
		"name":             {aclParams.Name},
		"config.whitelist": {aclParams.WhiteList},
	}
	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), PluginsPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to set up acl -- %s", err.Error())
		s.loggingClient.Error(e)
		return err
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up acl -- %s", err.Error())
		s.loggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		s.loggingClient.Info("acl set up successfully")
		break
	default:
		e := fmt.Sprintf("failed to set up acl with errorcode %d", resp.StatusCode)
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) initAuthMethod(name string, ttl int) error {
	s.loggingClient.Info(fmt.Sprintf("selected authetication method as %s.", name))
	switch name {
	case "jwt":
		return s.initJWTAuth()
	case "oauth2":
		return s.initOAuth2(ttl)
	default:
		return fmt.Errorf("unsupported authetication method: %s", name)
	}
}

func (s *Service) initJWTAuth() error {
	formVals := url.Values{
		"name": {"jwt"},
	}
	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), PluginsPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to create jwt auth request -- %s", err.Error())
		s.loggingClient.Error(e)
		return err
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up jwt authentication -- %s", err.Error())
		s.loggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		s.loggingClient.Info("successful to set up jwt authentication")
		break
	default:
		e := fmt.Sprintf("failed to set up jwt authentication with errorcode %d", resp.StatusCode)
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) initOAuth2(ttl int) error {
	oauth2Params := &KongOAuth2Plugin{
		Name:                    "oauth2",
		Scope:                   OAuth2Scopes,
		MandatoryScope:          "true",
		EnableClientCredentials: "true",
		EnableGlobalCredentials: "true",
		TokenTTL:                ttl,
	}
	//Again, the type above is largely useless but the struct tags indicate the field names below so I left it.
	formVals := url.Values{
		"name":                             {oauth2Params.Name},
		"config.scopes":                    {oauth2Params.Scope},
		"config.mandatory_scope":           {oauth2Params.MandatoryScope},
		"config.enable_client_credentials": {oauth2Params.EnableClientCredentials},
		"config.global_credentials":        {oauth2Params.EnableGlobalCredentials},
		"config.refresh_token_ttl":         {strconv.Itoa(oauth2Params.TokenTTL)},
	}
	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), PluginsPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to create oauth2 request -- %s", err.Error())
		s.loggingClient.Error(e)
		return err
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up oauth2 authentication -- %s", err.Error())
		s.loggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		s.loggingClient.Info("successful to set up oauth2 authentication")
		break
	default:
		e := fmt.Sprintf("failed to set up oauth2 authentication with errorcode %d", resp.StatusCode)
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) getSvcIDs(path string) (DataCollect, error) {
	collection := DataCollect{}

	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), path}
	req, err := http.NewRequest(http.MethodGet, strings.Join(tokens, "/"), nil)
	if err != nil {
		e := fmt.Sprintf("failed to create service list request -- %s", err.Error())
		s.loggingClient.Error(e)
		return collection, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get list of %s with error %s", path, err.Error())
		s.loggingClient.Error(err.Error())
		return collection, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if err = json.NewDecoder(resp.Body).Decode(&collection); err != nil {
			return collection, err
		}
		break
	default:
		e := fmt.Sprintf("failed to get list of %s with HTTP error code %d", path, resp.StatusCode)
		s.loggingClient.Error(e)
		return collection, errors.New(e)
	}
	return collection, nil
}
