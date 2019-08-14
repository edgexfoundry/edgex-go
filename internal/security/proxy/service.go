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
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

type Service struct {
	client internal.HttpCaller
}

func NewService(r internal.HttpCaller) Service {
	return Service{
		client: r,
	}
}

func (s *Service) CheckProxyServiceStatus() error {
	return s.checkServiceStatus(Configuration.KongURL.GetProxyBaseURL())
}

func (s *Service) CheckSecretServiceStatus() error {
	return s.checkServiceStatus(Configuration.SecretService.GetSecretSvcBaseURL())
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
		LoggingClient.Info(fmt.Sprintf("the service on %s is up successfully", path))
		break
	default:
		err = fmt.Errorf("unexpected http status %v %s", resp.StatusCode, path)
		LoggingClient.Error(err.Error())
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
			r := &Resource{c.ID, s.client}
			err = r.Remove(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Init(cert CertificateLoader) error {
	err := s.postCert(cert)
	if err != nil {
		return err
	}

	services := config.ListDefaultServices()
	//TODO: There are specific other services being routed through Kong. This should probably be driven through configuration
	//      For now, I'm putting this inline
	services["edgex-support-rulesengine"] = "Rulesengine"
	services["edgex-device-virtual"] = "Virtualdevice"

	for _, name := range services {
		params := Configuration.Clients[name]
		serviceParams := &KongService{
			Name:     name,
			Host:     params.Host,
			Port:     params.Port,
			Protocol: params.Protocol,
		}

		err := s.initKongService(serviceParams)
		if err != nil {
			return err
		}

		routeParams := &KongRoute{
			Paths: []string{"/" + name},
			Name:  name,
		}
		err = s.initKongRoutes(routeParams, name)
		if err != nil {
			return err
		}
	}

	err = s.initAuthMethod(Configuration.KongAuth.Name, Configuration.KongAuth.TokenTTL)
	if err != nil {
		return err
	}

	err = s.initACL(Configuration.KongACL.Name, Configuration.KongACL.WhiteList)
	if err != nil {
		return err
	}

	LoggingClient.Info("finishing initialization for reverse proxy")
	return nil
}

func (s *Service) postCert(cert CertificateLoader) error {
	cp, err := cert.Load()

	if err != nil {
		return err
	}

	body := &CertInfo{
		Cert: cp.Cert,
		Key:  cp.Key,
		Snis: Configuration.SecretService.SNIS,
	}
	LoggingClient.Debug("trying to upload cert to proxy server")
	data, err := json.Marshal(body)
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), CertificatesPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(string(data)))
	if err != nil {
		LoggingClient.Error("failed to create upload cert request -- %s", err.Error())
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		LoggingClient.Error("failed to upload cert to proxy server with error %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info("successfully added certificate to the reverse proxy")
		break
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		e := fmt.Sprintf("failed to add certificate with errorcode %d, error %s", resp.StatusCode, string(b))
		LoggingClient.Error(e)
		return errors.New(e)
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
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), ServicesPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		return fmt.Errorf("failed to construct http POST form request: %s %s", service.Name, err.Error())
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up proxy service: %s %s", service.Name, err.Error())
		LoggingClient.Error(e)
		return errors.New(e)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		LoggingClient.Info(fmt.Sprintf("successful to set up proxy service for %s", service.Name))
		break
	case http.StatusConflict:
		LoggingClient.Info(fmt.Sprintf("proxy service for %s has been set up", service.Name))
		break
	default:
		err = fmt.Errorf("proxy service for %s returned status %d", service.Name, resp.StatusCode)
		LoggingClient.Error(err.Error())
		return err
	}
	return nil
}

func (s *Service) initKongRoutes(r *KongRoute, name string) error {
	data, err := json.Marshal(r)
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), ServicesPath, name, "routes"}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(string(data)))
	if err != nil {
		e := fmt.Sprintf("failed to set up routes for %s with error %s", name, err.Error())
		LoggingClient.Error(e)
		return err
	}
	req.Header.Add(clients.ContentType, clients.ContentTypeJSON)

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up routes for %s with error %s", name, err.Error())
		LoggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info(fmt.Sprintf("successful to set up route for %s", name))
		break
	default:
		e := fmt.Sprintf("failed to set up route for %s with error %s", name, resp.Status)
		LoggingClient.Error(e)
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
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), PluginsPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to set up acl -- %s", err.Error())
		LoggingClient.Error(e)
		return err
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up acl -- %s", err.Error())
		LoggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info("acl set up successfully")
		break
	default:
		e := fmt.Sprintf("failed to set up acl with errorcode %d", resp.StatusCode)
		LoggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) initAuthMethod(name string, ttl int) error {
	LoggingClient.Info(fmt.Sprintf("selected authetication method as %s.", name))
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
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), PluginsPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to create jwt auth request -- %s", err.Error())
		LoggingClient.Error(e)
		return err
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up jwt authentication -- %s", err.Error())
		LoggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info("successful to set up jwt authentication")
		break
	default:
		e := fmt.Sprintf("failed to set up jwt authentication with errorcode %d", resp.StatusCode)
		LoggingClient.Error(e)
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
	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), PluginsPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to create oauth2 request -- %s", err.Error())
		LoggingClient.Error(e)
		return err
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up oauth2 authentication -- %s", err.Error())
		LoggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info("successful to set up oauth2 authentication")
		break
	default:
		e := fmt.Sprintf("failed to set up oauth2 authentication with errorcode %d", resp.StatusCode)
		LoggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) getSvcIDs(path string) (DataCollect, error) {
	collection := DataCollect{}

	tokens := []string{Configuration.KongURL.GetProxyBaseURL(), path}
	req, err := http.NewRequest(http.MethodGet, strings.Join(tokens, "/"), nil)
	if err != nil {
		e := fmt.Sprintf("failed to create service list request -- %s", err.Error())
		LoggingClient.Error(e)
		return collection, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get list of %s with error %s", path, err.Error())
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(e)
		return collection, errors.New(e)
	}
	return collection, nil
}
