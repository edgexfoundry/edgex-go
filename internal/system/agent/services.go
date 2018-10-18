/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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

package agent

import (
	"fmt"
	"time"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"github.com/edgexfoundry/edgex-go/internal"
)

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
)

func InvokeOperation(action string, services []string, params []string) bool {

	// Loop through requested operation, along with respectively-supplied parameters.
	for _, service := range services {
		LoggingClient.Info(fmt.Sprintf("About to {%v} the service {%v} with parameters {%v}.", action, service, params))

		switch action {

		case STOP:
			switch service {
			case internal.SupportNotificationsServiceKey:
				ec.StopService(internal.SupportNotificationsServiceKey, params[0])
				break

			case internal.CoreDataServiceKey:
				ec.StopService(internal.CoreDataServiceKey, params[0])
				break

			case internal.CoreCommandServiceKey:
				ec.StopService(internal.CoreCommandServiceKey, params[0])
				break

			case internal.CoreMetaDataServiceKey:
				ec.StopService(internal.CoreMetaDataServiceKey, params[0])
				break

			case internal.ExportClientServiceKey:
				ec.StopService(internal.ExportClientServiceKey, params[0])
				break

			case internal.ExportDistroServiceKey:
				ec.StopService(internal.ExportDistroServiceKey, params[0])
				break

			case internal.SupportLoggingServiceKey:
				ec.StopService(internal.SupportLoggingServiceKey, params[0])
				break

			case internal.ConfigSeedServiceKey:
				ec.StopService(internal.ConfigSeedServiceKey, params[0])
				break

			default:
				LoggingClient.Info(fmt.Sprintf(">> Unknown service: %v", service))
				break
			}

		case START:
			switch service {
			case internal.SupportNotificationsServiceKey:
				executor.StartDockerContainerCompose(internal.SupportNotificationsServiceKey)
				break

			case internal.CoreDataServiceKey:
				executor.StartDockerContainerCompose(internal.CoreDataServiceKey)
				break

			case internal.CoreMetaDataServiceKey:
				executor.StartDockerContainerCompose(internal.CoreMetaDataServiceKey)
				break

			case internal.CoreCommandServiceKey:
				executor.StartDockerContainerCompose(internal.CoreCommandServiceKey)
				break

			case internal.ExportClientServiceKey:
				executor.StartDockerContainerCompose(internal.ExportClientServiceKey)
				break

			case internal.ExportDistroServiceKey:
				executor.StartDockerContainerCompose(internal.ExportDistroServiceKey)
				break

			case internal.SupportLoggingServiceKey:
				executor.StartDockerContainerCompose(internal.SupportLoggingServiceKey)
				break

			case internal.ConfigSeedServiceKey:
				executor.StartDockerContainerCompose(internal.ConfigSeedServiceKey)
				break

			default:
				LoggingClient.Info(fmt.Sprintf(">> Unknown service: %v", service))
				break
			}

		case RESTART:
			switch service {
			case internal.SupportNotificationsServiceKey:
				ec.StopService(internal.SupportNotificationsServiceKey, params[0])
				time.Sleep(time.Second * time.Duration(1))
				executor.StartDockerContainerCompose(internal.SupportNotificationsServiceKey)
				break

			case internal.CoreDataServiceKey:
				ec.StopService(internal.CoreDataServiceKey, params[0])
				time.Sleep(time.Second * time.Duration(1))
				executor.StartDockerContainerCompose(internal.CoreDataServiceKey)
				break

			case internal.CoreCommandServiceKey:
				ec.StopService(internal.CoreCommandServiceKey, params[0])
				time.Sleep(time.Second * time.Duration(1))
				executor.StartDockerContainerCompose(internal.CoreCommandServiceKey)
				break

			case internal.CoreMetaDataServiceKey:
				ec.StopService(internal.CoreMetaDataServiceKey, params[0])
				time.Sleep(time.Second * time.Duration(1))
				executor.StartDockerContainerCompose(internal.CoreMetaDataServiceKey)
				break

			case internal.ExportClientServiceKey:
				ec.StopService(internal.ExportClientServiceKey, params[0])
				time.Sleep(time.Second * time.Duration(1))
				executor.StartDockerContainerCompose(internal.ExportClientServiceKey)
				break

			case internal.ExportDistroServiceKey:
				ec.StopService(internal.ExportDistroServiceKey, params[0])
				time.Sleep(time.Second * time.Duration(1))
				executor.StartDockerContainerCompose(internal.ExportDistroServiceKey)
				break

			case internal.SupportLoggingServiceKey:
				ec.StopService(internal.SupportLoggingServiceKey, params[0])
				time.Sleep(time.Second * time.Duration(1))
				executor.StartDockerContainerCompose(internal.SupportLoggingServiceKey)
				break

			case internal.ConfigSeedServiceKey:
				ec.StopService(internal.ConfigSeedServiceKey, params[0])
				time.Sleep(time.Second * time.Duration(1))
				executor.StartDockerContainerCompose(internal.ConfigSeedServiceKey)
				break

			default:
				LoggingClient.Info(fmt.Sprintf(">> Unknown service: %v", service))
				break
			}
		}
	}
	return true
}

func getConfig(services []string) (ConfigRespMap, error) {

	c := ConfigRespMap{}
	c.Configuration = map[string]interface{}{}

	// Loop through requested actions, along with respectively-supplied parameters.
	for _, service := range services {

		c.Configuration[service] = ""

		switch service {

		case internal.SupportNotificationsServiceKey:

			responseJSON, err := nc.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreCommandServiceKey:
			responseJSON, err := mcc.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreDataServiceKey:
			c.Configuration[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its configuration data...", internal.CoreDataServiceKey))
			break

		case internal.CoreMetaDataServiceKey:
			c.Configuration[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its configuration data...", internal.CoreMetaDataServiceKey))
			break

		case internal.ExportClientServiceKey:
			c.Configuration[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its configuration data...", internal.ExportClientServiceKey))
			break

		case internal.ExportDistroServiceKey:
			c.Configuration[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its configuration data...", internal.ExportDistroServiceKey))
			break

		case internal.SupportLoggingServiceKey:
			c.Configuration[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its configuration data...", internal.SupportLoggingServiceKey))
			break

		default:
			LoggingClient.Warn(fmt.Sprintf(">> Unknown service: %v", service))
			break
		}
	}
	return c, nil
}

func getMetrics(services []string) (MetricsRespMap, error) {

	m := MetricsRespMap{}
	m.Metrics = map[string]interface{}{}

	// Loop through requested actions, along with respectively-supplied parameters.
	for _, service := range services {

		m.Metrics[service] = ""

		switch service {

		case internal.SupportNotificationsServiceKey:
			responseJSON, err := nc.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreCommandServiceKey:
			responseJSON, err := mcc.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreDataServiceKey:
			m.Metrics[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its metrics data...", internal.CoreDataServiceKey))
			break

		case internal.CoreMetaDataServiceKey:
			m.Metrics[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its metrics data...", internal.CoreMetaDataServiceKey))
			break

		case internal.ExportClientServiceKey:
			m.Metrics[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its metrics data...", internal.ExportClientServiceKey))
			break

		case internal.ExportDistroServiceKey:
			m.Metrics[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its metrics data...", internal.ExportDistroServiceKey))
			break

		case internal.SupportLoggingServiceKey:
			m.Metrics[service] = "N/A"
			LoggingClient.Info(fmt.Sprintf("The micro-service {%v} currently does not support an endpoint for providing its metrics data...", internal.SupportLoggingServiceKey))
			break

		default:
			LoggingClient.Warn(fmt.Sprintf(">> Unknown service: %v", service))
			break
		}
	}
	return m, nil
}
