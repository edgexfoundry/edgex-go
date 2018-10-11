/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package agent

import (
	"sync"
	"time"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients/notifications"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

// Global variables
var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient
var Manifest *ManifestStruct
var Conf = &ConfigurationStruct{}
var nc notifications.ClientForNotifications
var Ec interfaces.ExecutorClient

func Retry(useConsul bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(useProfile)
			if err != nil {
				ch <- err
				if !useConsul {
					//Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			} else {
				// Initialize notificationsClient based on configuration
				initializeClients(useConsul)
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.SystemManagementAgentServiceKey, Configuration.EnableRemoteLogging, logTarget)
			}
		}

		// Exit the loop if the dependencies have been satisfied.
		if Configuration != nil {
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}

	until = time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		// The SMA-managed services are bootstrapped by the SMA.
		// Read the SMA's TOML manifest file, which which specifies details for those services.
		if Manifest == nil {
			Manifest, err = initializePerManifest(useProfile)
			if err != nil {
				ch <- err
				if !useConsul {
					//Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			}
		}
		// Exit the loop if the dependencies have been satisfied.
		if Manifest != nil {
			Ec, _ = newExecutorClient(Manifest.OperationsType)
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func newExecutorClient(operationsType string) (interfaces.ExecutorClient, error) {

	// TODO: The abstraction which should be accessed via a global var.
	switch operationsType {
	case "os":
		return &executor.ExecuteOs{}, nil
	case "docker":
		return &executor.ExecuteDocker{}, nil
	default:
		return nil, nil
	}
}

func Init() bool {
	if Configuration == nil {
		return false
	}
	return true
}

func initializeConfiguration(useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain ConsulHost/Port
	err := config.LoadFromFile(useProfile, Conf)
	if err != nil {
		return nil, err
	}

	return Conf, nil
}

func initializePerManifest(useProfile string) (*ManifestStruct, error) {
	// Populate in-memory store with data from the SMA's TOML manifest file.
	man := &ManifestStruct{}
	err := LoadFromFile(useProfile, man)
	if err != nil {
		return nil, err
	}

	return man, nil
}

func setLoggingTarget() string {
	logTarget := Configuration.LoggingRemoteURL
	if !Configuration.EnableRemoteLogging {
		return Configuration.LoggingFile
	}
	return logTarget
}

func initializeClients(useConsul bool) {
	// Create notification client
	params := types.EndpointParams{
		ServiceKey:  internal.SupportNotificationsServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Notifications"].Url(),
		Interval: Configuration.Service.ClientMonitor,
	}

	nc = notifications.NewNotificationsClient(params, startup.Endpoint{})
}
