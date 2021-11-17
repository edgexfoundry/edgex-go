//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package types

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const DefaultProtocol = "http"

type GetAccessTokenCallback func() (string, error)

// ServiceConfig defines the information need to connect to the Configuration service and optionally register the service
// for discovery and health checks
type ServiceConfig struct {
	// The Protocol that should be used to connect to the Configuration service. HTTP is used if not set.
	Protocol string
	// Host is the hostname or IP address of the Configuration service
	Host string
	// Port is the HTTP port of the Configuration service
	Port int
	// Type is the implementation type of the Configuration service, i.e. consul
	Type string
	// BasePath is the base path with in the Configuration service where the your service's configuration is stored
	BasePath string
	// AccessToken is the token that is used to access the service configuration
	AccessToken string
	// GetAccessToken is a callback function that retrieves a new Access Token.
	// This callback is used when a '403 Forbidden' status is received from any call to the configuration provider service.
	GetAccessToken GetAccessTokenCallback
}

//
// A few helper functions for building URLs.
//

func (config ServiceConfig) GetUrl() string {
	return fmt.Sprintf("%s://%s:%v", config.GetProtocol(), config.Host, config.Port)
}

func (config *ServiceConfig) GetProtocol() string {
	if config.Protocol == "" {
		return "http"
	}

	return config.Protocol
}

func (config *ServiceConfig) PopulateFromUrl(providerUrl string) error {
	url, err := url.Parse(providerUrl)
	if err != nil {
		return fmt.Errorf("the format of Provider URL is incorrect (%s): %s", providerUrl, err.Error())
	}

	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return fmt.Errorf("the port from Provider URL is incorrect (%s): %s", providerUrl, err.Error())
	}

	config.Host = url.Hostname()
	config.Port = port

	typeAndProtocol := strings.Split(url.Scheme, ".")

	// TODO: Enforce both Type and Protocol present for release V2.0.0
	// Support for default protocol is for backwards compatibility with Fuji Device Services.
	switch len(typeAndProtocol) {
	case 1:
		config.Type = typeAndProtocol[0]
		config.Protocol = DefaultProtocol
	case 2:
		config.Type = typeAndProtocol[0]
		config.Protocol = typeAndProtocol[1]
	default:
		return fmt.Errorf("the Type and Protocol spec from Provider URL is incorrect: %s", providerUrl)
	}

	return nil
}
