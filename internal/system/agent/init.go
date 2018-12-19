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
	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/pkg/clients/general"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

// Global variables
var Configuration *interfaces.ConfigurationStruct
var Conf = &interfaces.ConfigurationStruct{}
var ec interfaces.ExecutorClient
var clients map[string]general.GeneralClient

func Retry(useConsul bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		// When looping, only handle configuration if it hasn't already been set.
		// Note, too, that the SMA-managed services are bootstrapped by the SMA.
		// Read in those setting, too, which specifies details for those services
		// (Those setting were _previously_ to be found in a now-defunct TOML manifest file).
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
				logs.BuildLoggingClient(Configuration, logTarget)
			}
		}

		// Exit the loop if the dependencies have been satisfied.
		if Configuration != nil {
			ec, _ = newExecutorClient(Configuration.OperationsType)
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
	case "snap":
		return &executor.ExecuteSnap{}, nil
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

func initializeConfiguration(useProfile string) (*interfaces.ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain ConsulHost/Port
	err := config.LoadFromFile(useProfile, Conf)
	if err != nil {
		return nil, err
	}

	return Conf, nil
}

func setLoggingTarget() string {
	logTarget := Configuration.LoggingRemoteURL
	if !Configuration.EnableRemoteLogging {
		return Configuration.LoggingFile
	}
	return logTarget
}

func initializeClients(useConsul bool) {
	services := map[string]string{
		internal.SupportNotificationsServiceKey: "Notifications",
		internal.CoreCommandServiceKey: "Command",
		internal.CoreDataServiceKey: "CoreData",
		internal.CoreMetaDataServiceKey: "Metadata",
		internal.ExportClientServiceKey: "Export",
		internal.ExportDistroServiceKey: "Distro",
		internal.SupportLoggingServiceKey: "Logging",
		internal.SupportSchedulerServiceKey: "Scheduler",
	}

	clients = make(map[string]general.GeneralClient)

	for serviceKey, serviceName := range services {
		params := types.EndpointParams{
			ServiceKey:  serviceKey,
			Path:        "/",
			UseRegistry: useConsul,
			Url:         Configuration.Clients[serviceName].Url(),
			Interval:    internal.ClientMonitorDefault,
		}
		clients[serviceKey] = general.NewGeneralClient(params, startup.Endpoint{})
	}
}
