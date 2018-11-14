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
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/pkg/clients/general"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

// Global variables
var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient
var Conf = &ConfigurationStruct{}

// executorClient is the empty interface so that we may type cast it
// to whatever operation we need it to do at runtime
var executorClient interface{}
var gccc general.GeneralClient
var gccd general.GeneralClient
var gccm general.GeneralClient
var gcec general.GeneralClient
var gced general.GeneralClient
var gcsl general.GeneralClient
var gcsn general.GeneralClient
var gcss general.GeneralClient

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
				LoggingClient = logger.NewClient(internal.SystemManagementAgentServiceKey, Configuration.EnableRemoteLogging, logTarget, Configuration.LoggingLevel)
			}
		}

		// Exit the loop if the dependencies have been satisfied.
		if Configuration != nil {
			executorClient, err = makeExecutorImplementation(Configuration)
			if err != nil {
				ch <- err
				close(ch)
				wait.Done()
				return
			}
			// need to also pass in the docker compose url from theconfiguration struct
			// if we're using docker
			// TODO fix this abstraction so we don't have to type cast the interface into a type
			if Configuration.OperationsType == "docker" {
				if dockerExecutor, ok := executorClient.(*executor.ExecuteDocker); ok {
					dockerExecutor.ComposeURL = Configuration.ComposeUrl
				}
			}
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

// makeExecutorImplementation returns an executor implementation that can be used
// to start/stop/restart services depending on what the implementation
// supports
func makeExecutorImplementation(config *ConfigurationStruct) (interface{}, error) {
	switch config.OperationsType {
	case "os":
		return &executor.ExecuteOs{}, nil
	case "docker":
		return &executor.ExecuteDocker{
			ComposeURL: config.ComposeUrl,
		}, nil
	case "snap":
		return &executor.ExecuteSnap{}, nil
	case "custom":
		return &executor.CustomProgram{
			Program: config.CustomExecutorProgram,
		}, nil
	default:
		return nil, fmt.Errorf("operation type %s not supported", config.OperationsType)
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

func setLoggingTarget() string {
	logTarget := Configuration.LoggingRemoteURL
	if !Configuration.EnableRemoteLogging {
		return Configuration.LoggingFile
	}
	return logTarget
}

func initializeClients(useConsul bool) {
	// Create support-notifications client.
	paramsNotifications := types.EndpointParams{
		ServiceKey:  internal.SupportNotificationsServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Notifications"].Url(),
		Interval:    internal.ClientMonitorDefault,
	}
	gcsn = general.NewGeneralClient(paramsNotifications, startup.Endpoint{})

	// Create core-command client.
	paramsCoreCommand := types.EndpointParams{
		ServiceKey:  internal.CoreCommandServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Command"].Url(),
		Interval:    internal.ClientMonitorDefault,
	}
	gccc = general.NewGeneralClient(paramsCoreCommand, startup.Endpoint{})

	// Create core-data client.
	paramsCoreData := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["CoreData"].Url(),
		Interval:    internal.ClientMonitorDefault,
	}
	gccd = general.NewGeneralClient(paramsCoreData, startup.Endpoint{})

	// Create core-metadata client.
	paramsCoreMetadata := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Metadata"].Url(),
		Interval:    internal.ClientMonitorDefault,
	}
	gccm = general.NewGeneralClient(paramsCoreMetadata, startup.Endpoint{})

	// Create export-client client.
	paramsExportClient := types.EndpointParams{
		ServiceKey:  internal.ExportClientServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Export"].Url(),
		Interval:    internal.ClientMonitorDefault,
	}
	gcec = general.NewGeneralClient(paramsExportClient, startup.Endpoint{})

	// Create export-distro client.
	paramsExportDistro := types.EndpointParams{
		ServiceKey:  internal.ExportDistroServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Distro"].Url(),
		Interval:    internal.ClientMonitorDefault,
	}
	gced = general.NewGeneralClient(paramsExportDistro, startup.Endpoint{})

	// Create support-logging client.
	paramsSupportLogging := types.EndpointParams{
		ServiceKey:  internal.SupportLoggingServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Logging"].Url(),
		Interval:    internal.ClientMonitorDefault,
	}
	gcsl = general.NewGeneralClient(paramsSupportLogging, startup.Endpoint{})

	// Create support-scheduler client.
	paramsSupportScheduler := types.EndpointParams{
		ServiceKey:  internal.SupportSchedulerServiceKey,
		Path:        "/",
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Scheduler"].Url(),
		Interval:    internal.ClientMonitorDefault,
	}
	gcss = general.NewGeneralClient(paramsSupportScheduler, startup.Endpoint{})
}
