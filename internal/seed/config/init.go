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
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	consulapi "github.com/hashicorp/consul/api"
	"io/ioutil"
)

// Global variables
var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient
var Registry *consulapi.Client

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
				LoggingClient = logger.NewClient(internal.ConfigSeedServiceKey, Configuration.EnableRemoteLogging, logTarget)
			}
		}
		//Check to verify consul connectivity
		if Registry == nil {
			Registry, err = initConsulClient()
			if err != nil {
				ch <- err
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

func initConsulClient() (*consulapi.Client, error) {
	url := Configuration.ConsulProtocol + "://" + Configuration.ConsulHost + ":" + strconv.Itoa(Configuration.ConsulPort)

	resp, err := http.Get(url + consulStatusPath)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Connect to the Consul Agent
		cfg := consulapi.DefaultConfig()
		cfg.Address = url

		r, err := consulapi.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	body, err := getBody(resp)
	if err != nil {
		return nil, err
	}
	return nil, types.NewErrServiceClient(resp.StatusCode, body)
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
