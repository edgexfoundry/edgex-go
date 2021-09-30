/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *******************************************************************************/

package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/models"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
)

const (
	requestTransformerPlugin = "request-transformer"
	consulTokenHeader        = "X-Consul-Token" // nolint:gosec
)

// addConsulHeader is to enable request transformer plugin on a service
// with the given configuration
func (s *Service) addConsulTokenHeaderTo(kongService *models.KongService) error {
	// first to get the pluginID for the service
	// create one if not found
	servicePluginID, err := s.getServicePluginIDBy(kongService.Name, requestTransformerPlugin)
	if err != nil {
		return fmt.Errorf("cannot get service plugin ID: %w", err)
	}

	if len(servicePluginID) > 0 {
		// In order to make sure the Consul header always match with the latest one from the file
		// we want to delete the existing service plugin and re-create one as both Kong and Consul are stateful
		if err := s.deleteServicePlugin(kongService.Name, servicePluginID); err != nil {
			return fmt.Errorf("cannot delete service plugin ID: %w", err)
		}
	}

	accessToken, err := s.retrieveConsulAccessToken()
	if err != nil {
		return fmt.Errorf("cannot retrieve Consul access token: %w", err)
	}

	if len(accessToken) == 0 {
		return fmt.Errorf("retrieved access token is empty")
	}

	return s.createConsulTokenHeaderForServicePlugin(kongService.Name, requestTransformerPlugin, accessToken)
}

func (s *Service) getServicePluginIDBy(serviceName, pluginName string) (string, error) {
	servicePluginPath := path.Join(ServicesPath, serviceName, PluginsPath)
	servicePluginURL := strings.Join([]string{s.configuration.KongURL.GetProxyBaseURL(), servicePluginPath}, "/")

	req, err := http.NewRequest(http.MethodGet, servicePluginURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create get request for %s: %w", servicePluginURL, err)
	}

	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+s.bearerToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get plugin list for service %s: %w", serviceName, err)
	}

	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body to get plugins: %w", err)
	}

	type ServicePluginInfo struct {
		PluginID   string `json:"id"`
		PluginName string `json:"name"`
	}

	type KongServicePlugins struct {
		Entries []ServicePluginInfo `json:"data"`
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var plugins KongServicePlugins
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&plugins); err != nil {
			return "", fmt.Errorf("unable to parse response from get service plugins: %w", err)
		}

		for _, data := range plugins.Entries {
			if data.PluginName == pluginName {
				s.loggingClient.Infof("successful to get service pluginID of plugin %s for service %s", pluginName, serviceName)

				return data.PluginID, nil
			}
		}

		// no plugin matching pluginName found
		return "", nil
	default:
		return "", fmt.Errorf("unable to get service plugins from Kong's API %s with status code %d: %s", servicePluginURL,
			resp.StatusCode, string(responseBody))
	}
}

func (s *Service) deleteServicePlugin(serviceName, pluginID string) error {
	deletePluginPath := path.Join(ServicesPath, serviceName, PluginsPath, pluginID)
	deleteServicePluginURL := strings.Join([]string{s.configuration.KongURL.GetProxyBaseURL(), deletePluginPath}, "/")

	req, err := http.NewRequest(http.MethodDelete, deleteServicePluginURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create delete request for %s: %w", deleteServicePluginURL, err)
	}

	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+s.bearerToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete service plugin for service %s: %w", serviceName, err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNoContent:
		s.loggingClient.Infof("successfully deleted service plugin for service %s", serviceName)

		return nil
	default:
		responseBody, err := io.ReadAll(resp.Body)
		var errMsg string
		if err != nil {
			errMsg = fmt.Sprintf("failed to read response body for delete service plugin: %v", err)
		} else {
			errMsg = string(responseBody)
		}

		return fmt.Errorf("unable to delete service plugin from Kong's API %s with status code %d: %s", deleteServicePluginURL,
			resp.StatusCode, errMsg)
	}
}

func (s *Service) retrieveConsulAccessToken() (string, error) {
	if len(s.configuration.AccessTokenFile) == 0 {
		return "", errors.New("required AccessTokenFile from configuration is empty")
	}

	tokenFileAbsPath, err := filepath.Abs(s.configuration.AccessTokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to convert tokenFile to absolute path %s: %v", s.configuration.AccessTokenFile, err)
	}

	if exists := helper.CheckIfFileExists(tokenFileAbsPath); !exists {
		return "", fmt.Errorf("access token file %s not found", tokenFileAbsPath)
	}

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenReader, err := fileOpener.OpenFileReader(tokenFileAbsPath, os.O_RDONLY, 0400)
	if err != nil {
		return "", fmt.Errorf("failed to open file reader: %v", err)
	}

	var accessToken map[string]interface{}
	if err := json.NewDecoder(tokenReader).Decode(&accessToken); err != nil {
		return "", fmt.Errorf("failed to parse access token JSON data: %v", err)
	}

	s.loggingClient.Infof("successfully retrieved access token from %s", tokenFileAbsPath)

	return fmt.Sprintf("%s", accessToken["SecretID"]), nil
}

func (s *Service) createConsulTokenHeaderForServicePlugin(serviceName, pluginName, accessToken string) error {
	servicePluginPath := path.Join(ServicesPath, serviceName, PluginsPath)
	servicePluginURL := strings.Join([]string{s.configuration.KongURL.GetProxyBaseURL(), servicePluginPath}, "/")

	form := url.Values{
		"name":               []string{pluginName},
		"config.add.headers": []string{consulTokenHeader + ":" + accessToken},
	}

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, servicePluginURL, strings.NewReader(formVal))
	if err != nil {
		return fmt.Errorf("failed to create POST request for %s: %w", servicePluginURL, err)
	}

	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+s.bearerToken)
	req.Header.Add(common.ContentType, URLEncodedForm)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create Consul token header for service %s: %w", serviceName, err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusCreated:
		s.loggingClient.Infof("successfully created Consul token header to service %s", serviceName)

		return nil
	default:
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body for creating Consul token header: %w", err)
		}

		return fmt.Errorf("unable to create Consul token header from Kong's API %s with status code %d: %s", servicePluginURL,
			resp.StatusCode, string(responseBody))
	}
}
