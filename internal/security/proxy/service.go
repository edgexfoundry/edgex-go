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
	"github.com/dghubble/sling"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"io/ioutil"
	"net/http"
)

type Service struct {
	client Requestor
}

func NewService(r Requestor) Service {
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
	req, err := sling.New().Get(path).Request()
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
	req, err := sling.New().Base(Configuration.KongURL.GetProxyBaseURL()).Post(CertificatesPath).BodyJSON(body).Request()
	if err != nil {
		LoggingClient.Error("failed to upload cert to proxy server with error %s", err.Error())
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
	req, err := sling.New().Base(Configuration.KongURL.GetProxyBaseURL()).Post(ServicesPath).BodyForm(service).Request()
	if err != nil {
		e := fmt.Sprintf("failed to set up proxy service for %s", service.Name)
		return errors.New(e)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up proxy service for %s", service.Name)
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
	routesubpath := "services/" + name + "/routes"
	req, err := sling.New().Base(Configuration.KongURL.GetProxyBaseURL()).Post(routesubpath).BodyJSON(r).Request()
	if err != nil {
		e := fmt.Sprintf("failed to set up routes for %s with error %s", name, err.Error())
		LoggingClient.Error(e)
		return err
	}
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
	req, err := sling.New().Base(Configuration.KongURL.GetProxyBaseURL()).Post(PluginsPath).BodyForm(aclParams).Request()
	if err != nil {
		e := fmt.Sprintf("failed to set up acl")
		LoggingClient.Error(e)
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up acl")
		LoggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		LoggingClient.Info("successful to set up acl")
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
	jwtParams := &KongJWTPlugin{
		Name: "jwt",
	}

	req, err := sling.New().Base(Configuration.KongURL.GetProxyBaseURL()).Post(PluginsPath).BodyForm(jwtParams).Request()
	if err != nil {
		e := fmt.Sprintf("failed to set up jwt authentication")
		LoggingClient.Error(e)
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up jwt authentication")
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

	req, err := sling.New().Base(Configuration.KongURL.GetProxyBaseURL()).Post(PluginsPath).BodyForm(oauth2Params).Request()
	if err != nil {
		e := fmt.Sprintf("failed to set up oauth2 authentication with error %s", err.Error())
		LoggingClient.Error(e)
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up oauth2 authentication with error %s", err.Error())
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

	req, err := sling.New().Get(Configuration.KongURL.GetProxyBaseURL()).Path(path).Request()
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
