/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/go-mod-registry"
	"github.com/edgexfoundry/go-mod-registry/pkg/factory"
)

// Global variables
var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient
var Registry registry.Client

// The purpose of Retry is different here than in other services. In this case, we use a retry in order
// to initialize the ConsulClient that will be used to write configuration information. Other services
// use Retry to read their information. Config-seed writes information.
func Retry(useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(useProfile)
			if err != nil {
				ch <- err
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.ConfigSeedServiceKey, Configuration.EnableRemoteLogging, logTarget, Configuration.LoggingLevel)
			}
		}
		//Check to verify consul connectivity
		if Registry == nil {
			Registry, err = initRegistryClient("")

			if err != nil {
				ch <- err
			}
		} else {
			if !Registry.IsAlive() {
				ch <- fmt.Errorf("Registry (%s) is not running", Configuration.Registry.Type)
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
	if Configuration != nil && Registry != nil {
		return true
	}
	return false
}

func initializeConfiguration(useProfile string) (*ConfigurationStruct, error) {
	conf := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func initRegistryClient(serviceKey string) (registry.Client, error) {
	registryConfig := registry.Config{
		Host:       Configuration.Registry.Host,
		Port:       Configuration.Registry.Port,
		Type:       Configuration.Registry.Type,
		ServiceKey: serviceKey,
	}
	registryClient, err := factory.NewRegistryClient(registryConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create New Registry: %v", err)
	}

	if !registryClient.IsAlive() {
		return nil, fmt.Errorf("registry is not available")

	}

	return registryClient, nil
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
