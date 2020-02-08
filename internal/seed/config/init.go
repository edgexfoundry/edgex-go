/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
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
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-configuration/configuration"
	"github.com/edgexfoundry/go-mod-configuration/pkg/types"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
)

// Global variables
var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient
var ConfigClient configuration.Client

// The purpose of Retry is different here than in other services. In this case, we use a retry in order
// to initialize the RegistryClient that will be used to write configuration information. Other services
// use Retry to read their information. Config-seed writes information.
func Retry(configDir, profileDir string, configProviderUrl string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(configDir, profileDir)
			if err != nil {
				ch <- err
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(clients.ConfigSeedServiceKey, Configuration.EnableRemoteLogging, logTarget, Configuration.LoggingLevel)
			}
		}
		//Check to verify Configuration Provider connectivity
		if ConfigClient == nil {
			ConfigClient, err = initConfigClient("", configProviderUrl)

			if err != nil {
				ch <- err
			}
		} else {
			if !ConfigClient.IsAlive() {
				ch <- fmt.Errorf("Configuration Provider (%s) is not running", configProviderUrl)
			} else {
				break
			}
		}

		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func Init() bool {
	if Configuration != nil && ConfigClient != nil {
		return true
	}
	return false
}

func initializeConfiguration(configDir, profileDir string) (*ConfigurationStruct, error) {
	conf := &ConfigurationStruct{}
	err := config.LoadFromFile(configDir, profileDir, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func initConfigClient(serviceKey string, configProviderUrl string) (configuration.Client, error) {
	if configProviderUrl == "" {
		return nil, fmt.Errorf("Configuation Provder URL must be specified via command line or environment variable")
	}

	providerConfig := types.ServiceConfig{}
	if err := providerConfig.PopulateFromUrl(configProviderUrl); err != nil {
		return nil, err
	}
	providerConfig.BasePath = serviceKey

	configClient, err := configuration.NewConfigurationClient(providerConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create New Registry: %v", err)
	}

	if !configClient.IsAlive() {
		return nil, fmt.Errorf("registry is not available")

	}

	return configClient, nil
}

// Helper method to get the body from the response after making the request
func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)

	return body, err
}

func setLoggingTarget() string {
	logTarget := Configuration.LoggingRemoteURL
	if !Configuration.EnableRemoteLogging {
		return Configuration.LoggingFile
	}
	return logTarget
}
