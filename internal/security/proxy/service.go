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
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/edgex-go/internal"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
)

type CertUploadErrorType int

const (
	CertExisting  CertUploadErrorType = 0
	InternalError CertUploadErrorType = 1

	// AddProxyRoutesEnv is the environment variable name for adding the additional Kong routes for app services
	AddProxyRoutesEnv = "ADD_PROXY_ROUTE"
)

type CertError struct {
	err    string
	reason CertUploadErrorType
}

func (ce *CertError) Error() string {
	return ce.err
}

type Service struct {
	client           internal.HttpCaller
	loggingClient    logger.LoggingClient
	configuration    *config.ConfigurationStruct
	additionalRoutes string
	routes           map[string]*KongRoute
}

func NewService(
	r internal.HttpCaller,
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct) Service {

	return Service{
		client:           r,
		loggingClient:    lc,
		configuration:    configuration,
		additionalRoutes: strings.TrimSpace(os.Getenv(AddProxyRoutesEnv)),
		routes:           make(map[string]*KongRoute),
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

func (s *Service) Init(cp bootstrapConfig.CertKeyPair) error {
	postErr := s.postCert(cp)
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

	addRoutesFromEnv, parseErr := s.parseAdditionalProxyRoutes()

	if parseErr != nil {
		s.loggingClient.Error(fmt.Sprintf(
			"failed to parse additional proxy Kong routes from env %s: %s",
			s.additionalRoutes, parseErr.Error()))
	}

	mergedClients := s.mergeRoutesWith(addRoutesFromEnv)

	for clientName, client := range mergedClients {
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

// parseAdditionalProxyRoutes is to parse out the value of env AddProxyRoutesEnv
// into key / value pairs of map [string]bootstrapConfig.ClientInfo
// where key is service name, and value is the service ClientInfo
// the env should contain the list of comma separated entries with the format of
// Name.URL to be considered well-formed
// Name is used as the key of service name
// URL should be well-formed as protocol://hostName:portNumber
// and is parsed into bootstrapConfig.ClientInfo structure if valid
// returns error if not valid
func (s *Service) parseAdditionalProxyRoutes() (map[string]bootstrapConfig.ClientInfo, error) {
	emptyMap := make(map[string]bootstrapConfig.ClientInfo)

	routesFromEnv := strings.Split(s.additionalRoutes, ",")

	additionalClientMap := make(map[string]bootstrapConfig.ClientInfo)
	for _, rt := range routesFromEnv {
		route := strings.TrimSpace(rt)
		// ignore the empty route
		if route == "" {
			continue
		}

		if !strings.Contains(route, ".") {
			// Invalid syntax for route, it should contain dot (.)
			return emptyMap, fmt.Errorf(
				"invalid syntax for defining additional kong route %s, it should contain dot . as separator", route)
		}

		// assume the routePair is in the format of serviceName.routeURL
		routePair := strings.SplitN(route, ".", 2)
		serviceName := strings.TrimSpace(routePair[0])
		routeURL := strings.TrimSpace(routePair[1])

		if serviceName == "" {
			// service name should not be empty
			return emptyMap, errors.New("service name for kong route should not be empty")
		}

		// sanity check to validate the well-formness of routeURL
		// and also parse out the protocol, hostname, and port number if it is good
		url, err := url.Parse(routeURL)
		if err != nil {
			return emptyMap, fmt.Errorf(
				"malformed route URL for additional kong route %s: %s", routeURL, err.Error())
		}
		hostName, port, err := net.SplitHostPort(url.Host)
		if err != nil {
			return emptyMap, fmt.Errorf(
				"malformed host in route URL for additional kong route %s: %s", url.Host, err.Error())
		}
		portNum, err := strconv.Atoi(port)
		if err != nil {
			return emptyMap, fmt.Errorf(
				"invalid port, expecting integer as port number for additional kong route %s: %s", port, err.Error())
		}

		clientInfo := bootstrapConfig.ClientInfo{
			Protocol: url.Scheme,
			Host:     hostName,
			Port:     portNum,
		}

		additionalClientMap[serviceName] = clientInfo
	}

	return additionalClientMap, nil
}

func (s *Service) mergeRoutesWith(additional map[string]bootstrapConfig.ClientInfo) map[string]bootstrapConfig.ClientInfo {
	// merging ignores the duplicate keys with the current internal map

	if len(additional) == 0 {
		return s.configuration.Clients
	}

	if len(s.configuration.Clients) == 0 {
		return additional
	}

	merged := make(map[string]bootstrapConfig.ClientInfo)
	for serviceName, client := range s.configuration.Clients {
		merged[serviceName] = client
	}

	for serviceName, client := range additional {
		_, exists := merged[serviceName]
		if exists {
			s.loggingClient.Warn(fmt.Sprintf(
				"attempting to add additional service name %s that already exists in the config. "+
					"Ignoring additional", serviceName))
			continue
		}
		merged[serviceName] = client
	}

	return merged
}

func (s *Service) postCert(cp bootstrapConfig.CertKeyPair) *CertError {
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
		s.routes[name] = r
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
