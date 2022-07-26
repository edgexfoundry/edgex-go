/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Corp.
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
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/models"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
)

type CertUploadErrorType int

const (
	CertExisting  CertUploadErrorType = 0
	InternalError CertUploadErrorType = 1

	// AddProxyRoutesEnv is the environment variable name for adding the additional Kong routes for app services
	AddProxyRoutesEnv = "ADD_PROXY_ROUTE"

	edgeXCoreConsulServiceKey = "core-consul"
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
	routes           map[string]*models.KongRoute
	bearerToken      string
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
		routes:           make(map[string]*models.KongRoute),
	}
}

func (s *Service) CheckProxyServiceStatus() error {
	return s.checkServiceStatus(s.configuration.KongURL.GetProxyStatusURL())
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
		s.loggingClient.Infof("the service on %s is up successfully", path)
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

func (s *Service) Init() error {

	// Read in JWT
	jwtBytes, err := os.ReadFile(s.configuration.KongAuth.JWTFile)
	if err != nil {
		return fmt.Errorf("failed to read config template: %w", err)
	}
	s.bearerToken = string(jwtBytes)

	addRoutesFromEnv, parseErr := s.parseAdditionalProxyRoutes()

	if parseErr != nil {
		s.loggingClient.Error(fmt.Sprintf(
			"failed to parse additional proxy Kong routes from env %s: %s",
			s.additionalRoutes, parseErr.Error()))
	}

	mergedRoutes := s.mergeRoutesWith(addRoutesFromEnv)

	for serviceKey, route := range mergedRoutes {

		var routeCopy = route // gosec -- memory aliasing in for loop
		err := s.initKongService(&routeCopy)
		if err != nil {
			return err
		}

		// if it is edgex-core-consul service; then we need to enable the request transformer plugin
		// in order to add the consul token header for that service
		// see details on https://docs.konghq.com/hub/kong-inc/request-transformer/#enabling-the-plugin-on-a-service
		if serviceKey == edgeXCoreConsulServiceKey {
			s.loggingClient.Infof("try to enable service plugin for %s", edgeXCoreConsulServiceKey)
			if err := s.addConsulTokenHeaderTo(&routeCopy); err != nil {
				s.loggingClient.Errorf("failed to enable service plugin for %s: %v", edgeXCoreConsulServiceKey, err)
				return err
			}

			s.loggingClient.Infof("service plugin for %s enabled", edgeXCoreConsulServiceKey)
		}

		routeParams := &models.KongRoute{
			Paths: []string{"/" + strings.ToLower(route.Name)},
			Name:  strings.ToLower(route.Name),
		}

		if err := s.initKongRoutes(routeParams, strings.ToLower(route.Name)); err != nil {
			return err
		}

		if s.configuration.CORSConfiguration.EnableCORS {
			err = s.initCORSRoutes(&s.configuration.CORSConfiguration, strings.ToLower(route.Name))
			if err != nil {
				return err
			}
		}

		var formVals url.Values

		switch s.configuration.KongAuth.Name {
		case "jwt":
			formVals = url.Values{
				"name": {"jwt"},
			}
		default:
			return fmt.Errorf("unsupported authetication method: %s", s.configuration.KongAuth.Name)
		}

		s.loggingClient.Infof("selected auth method %s", s.configuration.KongAuth.Name)

		if err := s.initRouteAuthentication(strings.ToLower(route.Name), formVals.Encode()); err != nil {
			return err
		}

	}

	s.loggingClient.Info("finishing initialization for reverse proxy")
	return nil
}

// parseAdditionalProxyRoutes is to parse out the value of env AddProxyRoutesEnv
// into key / value pairs of map [string]KongService
// where key is service name, and value is the service
// the env should contain the list of comma separated entries with the format of
// Name.URL to be considered well-formed
// Name is used as the key of service name
// URL should be well-formed as protocol://hostName:portNumber
// and is parsed into KongService structure if valid
// returns error if not valid
func (s *Service) parseAdditionalProxyRoutes() (map[string]models.KongService, error) {
	emptyMap := make(map[string]models.KongService)

	routesFromEnv := strings.Split(s.additionalRoutes, ",")

	additionalRouteMap := make(map[string]models.KongService)
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

		routeInfo := models.KongService{
			Name:     serviceName,
			Protocol: url.Scheme,
			Host:     hostName,
			Port:     portNum,
		}

		additionalRouteMap[serviceName] = routeInfo
	}

	return additionalRouteMap, nil
}

func (s *Service) mergeRoutesWith(additional map[string]models.KongService) map[string]models.KongService {
	// merging ignores the duplicate keys with the current internal map

	if len(additional) == 0 {
		return s.configuration.Routes
	}

	if len(s.configuration.Routes) == 0 {
		return additional
	}

	merged := make(map[string]models.KongService)
	for serviceName, route := range s.configuration.Routes {
		merged[serviceName] = route
	}

	for serviceName, route := range additional {
		_, exists := merged[serviceName]
		if exists {
			s.loggingClient.Warnf(
				"attempting to add additional service name %s that already exists in the config. "+
					"Ignoring additional", serviceName)
			continue
		}
		merged[serviceName] = route
	}

	return merged
}

func (s *Service) postCert(cp bootstrapConfig.CertKeyPair) *CertError {
	body := &models.CertInfo{
		Cert: cp.Cert,
		Key:  cp.Key,
		Snis: s.configuration.SNIS,
	}
	s.loggingClient.Debug("trying to upload cert to proxy server")
	data, err := json.Marshal(body)
	if err != nil {
		s.loggingClient.Error(err.Error())
		return &CertError{err.Error(), InternalError}
	}
	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), CertificatesPath}
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(string(data)))
	if err != nil {
		s.loggingClient.Errorf("failed to create upload cert request -- %s", err.Error())
		return &CertError{err.Error(), InternalError}
	}
	req.Header.Add(common.ContentType, common.ContentTypeJSON)
	resp, err := s.client.Do(req)
	if err != nil {
		s.loggingClient.Errorf("failed to upload cert to proxy server with error %s", err.Error())
		return &CertError{err.Error(), InternalError}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		s.loggingClient.Info("successfully added certificate to the reverse proxy")
	default:
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return &CertError{err.Error(), InternalError}
		}
		why := string(b)
		message := fmt.Sprintf("failed to add certificate with errorcode %d, error %s", resp.StatusCode, why)
		s.loggingClient.Error(message)

		if (resp.StatusCode == http.StatusBadRequest) && strings.Contains(why, "existing certificate") {
			message = "certificate already exists on reverse proxy"
		}
		return &CertError{message, CertExisting}
	}
	return nil
}

func (s *Service) initKongService(service *models.KongService) error {
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
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+s.bearerToken)
	req.Header.Add(common.ContentType, URLEncodedForm)

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up proxy service: %s %s", service.Name, err.Error())
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		s.loggingClient.Infof("successful to set up proxy service for `%s` at `%s:%d`", service.Name, service.Host, service.Port)
	case http.StatusConflict:
		s.loggingClient.Infof("proxy service for %s has been set up", service.Name)
	default:
		err = fmt.Errorf("proxy service for %s returned status %d", service.Name, resp.StatusCode)
		s.loggingClient.Error(err.Error())
		return err
	}
	return nil
}

func (s *Service) initKongRoutes(r *models.KongRoute, name string) error {
	data, err := json.Marshal(r)
	if err != nil {
		s.loggingClient.Error(err.Error())
		return err
	}
	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), ServicesPath, name, "routes"}

	// Create routes associated to a specific service
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(string(data)))
	if err != nil {
		e := fmt.Sprintf("failed to set up routes for %s with error %s", name, err.Error())
		s.loggingClient.Error(e)
		return err
	}
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+s.bearerToken)
	req.Header.Add(common.ContentType, common.ContentTypeJSON)

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
		s.loggingClient.Infof("successful to set up route for `%s` at `%v`", name, r.Paths)
	default:
		e := fmt.Sprintf("failed to set up route for %s with error %s", name, resp.Status)
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) initCORSRoutes(corsConfig *bootstrapConfig.CORSConfigurationInfo, name string) error {
	formVals := url.Values{
		"name":               {"cors"},
		"config.origins":     {corsConfig.CORSAllowedOrigin},
		"config.credentials": {strconv.FormatBool(corsConfig.CORSAllowCredentials)},
		"config.max_age":     {strconv.Itoa(corsConfig.CORSMaxAge)},
	}

	// Break out CORSAllowedMethods and CORSAllowedHeaders
	for _, method := range strings.Split(corsConfig.CORSAllowedMethods, ",") {
		formVals.Add("config.methods", strings.TrimSpace(method))
	}
	for _, header := range strings.Split(corsConfig.CORSAllowedHeaders, ",") {
		formVals.Add("config.headers", strings.TrimSpace(header))
	}
	for _, header := range strings.Split(corsConfig.CORSExposeHeaders, ",") {
		formVals.Add("config.exposed_headers", strings.TrimSpace(header))
	}

	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), RoutesPath, name, "plugins"}

	// Create routes associated to a specific service
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals.Encode()))
	if err != nil {
		e := fmt.Sprintf("failed to set up routes for %s with error %s", name, err.Error())
		s.loggingClient.Error(e)
		return err
	}
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+s.bearerToken)
	req.Header.Add(common.ContentType, URLEncodedForm)

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up CORS for %s with error %s", name, err.Error())
		s.loggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		s.loggingClient.Infof("successful to set up CORS at `%s`", name)
	default:
		e := fmt.Sprintf("failed to set up CORS for %s with error %s", name, resp.Status)
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) initRouteAuthentication(routeName string, formVals string) error {

	tokens := []string{s.configuration.KongURL.GetProxyBaseURL(), RoutesPath, routeName, PluginsPath}

	// Create routes associated to a specific service
	req, err := http.NewRequest(http.MethodPost, strings.Join(tokens, "/"), strings.NewReader(formVals))
	if err != nil {
		e := fmt.Sprintf("failed to form route authentication request for %s with error %s", routeName, err.Error())
		s.loggingClient.Error(e)
		return err
	}
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+s.bearerToken)
	req.Header.Add(common.ContentType, URLEncodedForm)

	resp, err := s.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to request route authentication for %s with error %s", routeName, err.Error())
		s.loggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		s.loggingClient.Infof("successful to set up route authentication for `%s`", routeName)
	default:
		body, _ := io.ReadAll(resp.Body)
		e := fmt.Sprintf("failed to set up route authentication for %s with error %s, %s", routeName, resp.Status, string(body))
		s.loggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}

func (s *Service) getSvcIDs(path string) (models.DataCollect, error) {
	collection := models.DataCollect{}

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
	default:
		e := fmt.Sprintf("failed to get list of %s with HTTP error code %d", path, resp.StatusCode)
		s.loggingClient.Error(e)
		return collection, errors.New(e)
	}
	return collection, nil
}
