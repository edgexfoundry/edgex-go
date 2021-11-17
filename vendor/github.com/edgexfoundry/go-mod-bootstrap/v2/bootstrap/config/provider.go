/*******************************************************************************
 * Copyright 2020 Intel Corp.
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
 *******************************************************************************/
package config

import (
	"github.com/edgexfoundry/go-mod-configuration/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/environment"
)

// ProviderInfo encapsulates the usage of the Configuration Provider information
type ProviderInfo struct {
	serviceConfig types.ServiceConfig
}

// NewProviderInfo creates a new ProviderInfo and initializes it
func NewProviderInfo(envVars *environment.Variables, providerUrl string) (*ProviderInfo, error) {
	var err error
	configProviderInfo := ProviderInfo{}

	// initialize config provider configuration for URL set in commandline options
	if providerUrl != "" {
		if err = configProviderInfo.serviceConfig.PopulateFromUrl(providerUrl); err != nil {
			return nil, err
		}
	}

	// override file-based configuration with Variables variables.
	configProviderInfo.serviceConfig, err = envVars.OverrideConfigProviderInfo(configProviderInfo.serviceConfig)
	if err != nil {
		return nil, err
	}

	return &configProviderInfo, nil
}

// UseProvider returns whether the Configuration Provider should be used or not.
func (config ProviderInfo) UseProvider() bool {
	return config.serviceConfig.Host != ""
}

// ServiceConfig returns service configuration for the Configuration Provider
func (config ProviderInfo) ServiceConfig() types.ServiceConfig {
	return config.serviceConfig
}
