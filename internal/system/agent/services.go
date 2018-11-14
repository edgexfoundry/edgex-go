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
	"errors"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
)

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
	DISABLE = "disable"
	ENABLE  = "enable"
)

// InvokeOperation attempts to take the specified action on as many
// of the services as it can, returning
func InvokeOperation(action string, services []string, params []string) error {

	allFailedSvcs := make(map[string]string)
	// make a map of the service names and use it as a set
	// to check for membership
	serviceNameSet := make(map[string]bool)
	for _, supportedService := range []string{
		// note that the sys-mgmt-agent is not here so that if
		// the sys-mgmt-agent gets a request to control itself that
		// can be handled specially
		internal.ConfigSeedServiceKey,
		internal.CoreCommandServiceKey,
		internal.CoreDataServiceKey,
		internal.CoreMetaDataServiceKey,
		internal.ExportClientServiceKey,
		internal.ExportDistroServiceKey,
		internal.SupportLoggingServiceKey,
		internal.SupportNotificationsServiceKey,
		internal.SupportSchedulerServiceKey,
	} {
		serviceNameSet[supportedService] = true
	}

	// Loop through requested operation, along with respectively-supplied parameters.
	for _, service := range services {
		LoggingClient.Info(fmt.Sprintf("About to {%v} the service {%v} ", action, service))

		// check if this request is for the sys-mgmt-agent itself, and if so issue warning and ignore
		if service == internal.SystemManagementAgentServiceKey {
			msg := "received request to manage sys-mgmt-agent, ignoring"
			LoggingClient.Error(msg)
			allFailedSvcs[service] = msg
			LoggingClient.Warn(msg)
			continue
		}

		// make sure this is a known service, otherwise log an error and continue
		if _, found := serviceNameSet[service]; !found {
			msg := fmt.Sprintf("unknown service \"%s\"", service)
			LoggingClient.Error(msg)
			allFailedSvcs[service] = msg
			continue
		}

		switch action {

		case STOP:
			if stopper, ok := executorClient.(interfaces.ServiceStopper); ok {
				err := stopper.Stop(service, params)
				if err != nil {
					msg := fmt.Sprintf("error stopping service \"%s\": %v", service, err)
					LoggingClient.Error(msg)
					allFailedSvcs[service] = msg
				}
			} else {
				msg := fmt.Sprintf("stopping not supported with specified executor of \"%s\"", Configuration.OperationsType)
				LoggingClient.Warn(msg)
				allFailedSvcs[service] = msg
			}

		case START:
			if starter, ok := executorClient.(interfaces.ServiceStarter); ok {
				err := starter.Start(service, params)
				if err != nil {
					msg := fmt.Sprintf("error starting service \"%s\": %v", service, err)
					LoggingClient.Error(msg)
					allFailedSvcs[service] = msg
				}
			} else {
				msg := fmt.Sprintf("starting not supported with specified executor of \"%s\"", Configuration.OperationsType)
				LoggingClient.Warn(msg)
				allFailedSvcs[service] = msg
			}

		case RESTART:
			if restarter, ok := executorClient.(interfaces.ServiceRestarter); ok {
				err := restarter.Restart(service, params)
				if err != nil {
					msg := fmt.Sprintf("error restarting service \"%s\": %v", service, err)
					LoggingClient.Error(msg)
					allFailedSvcs[service] = msg
				}
			} else {
				msg := fmt.Sprintf("restarting not supported with specified executor of \"%s\"", Configuration.OperationsType)
				LoggingClient.Warn(msg)
				allFailedSvcs[service] = msg
			}

		case ENABLE:
			if enabler, ok := executorClient.(interfaces.ServiceEnabler); ok {
				err := enabler.Enable(service, params)
				if err != nil {
					msg := fmt.Sprintf("error enabling service \"%s\": %v", service, err)
					LoggingClient.Error(msg)
					allFailedSvcs[service] = msg
				}
			} else {
				msg := fmt.Sprintf("enabling not supported with specified executor of \"%s\"", Configuration.OperationsType)
				LoggingClient.Warn(msg)
				allFailedSvcs[service] = msg
			}

		case DISABLE:
			if disabler, ok := executorClient.(interfaces.ServiceDisabler); ok {
				err := disabler.Disable(service, params)
				if err != nil {
					msg := fmt.Sprintf("error disabling service \"%s\": %v", service, err)
					LoggingClient.Error(msg)
					allFailedSvcs[service] = msg
				}
			} else {
				msg := fmt.Sprintf("disabling not supported with specified executor of \"%s\"", Configuration.OperationsType)
				LoggingClient.Warn(msg)
				allFailedSvcs[service] = msg
			}
		default:
			msg := fmt.Sprintf("unsupported operation \"%s\" on service \"%s\"", action, service)
			LoggingClient.Error(msg)
			allFailedSvcs[service] = msg
		}
	}

	if len(allFailedSvcs) != 0 {
		msg := ""
		for svc, failMsg := range allFailedSvcs {
			msg += fmt.Sprintf("failed to perform action \"%s\" on service %s because: %v\n", action, svc, failMsg)
		}
		return errors.New(msg)
	}

	return nil
}

func getConfig(services []string) (ConfigRespMap, error) {

	c := ConfigRespMap{}
	c.Configuration = map[string]interface{}{}

	// Loop through requested actions, along with respectively-supplied parameters.
	for _, service := range services {

		c.Configuration[service] = ""

		switch service {

		case internal.SupportNotificationsServiceKey:

			responseJSON, err := gcsn.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreCommandServiceKey:
			responseJSON, err := gccc.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreDataServiceKey:
			responseJSON, err := gccd.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreMetaDataServiceKey:
			responseJSON, err := gccm.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.ExportClientServiceKey:
			responseJSON, err := gcec.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.ExportDistroServiceKey:
			responseJSON, err := gced.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.SupportLoggingServiceKey:
			responseJSON, err := gcsl.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.SupportSchedulerServiceKey:
			responseJSON, err := gcss.FetchConfiguration()
			if err != nil {
				msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
				c.Configuration[service] = msg
				LoggingClient.Error(msg)
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
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
			responseJSON, err := gcsn.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreCommandServiceKey:
			responseJSON, err := gccc.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreDataServiceKey:
			responseJSON, err := gccd.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.CoreMetaDataServiceKey:
			responseJSON, err := gccm.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.ExportClientServiceKey:
			responseJSON, err := gcec.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.ExportDistroServiceKey:
			responseJSON, err := gced.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.SupportLoggingServiceKey:
			responseJSON, err := gcsl.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		case internal.SupportSchedulerServiceKey:
			responseJSON, err := gcss.FetchMetrics()
			if err != nil {
				msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
				m.Metrics[service] = msg
				LoggingClient.Error(msg)
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
			break

		default:
			LoggingClient.Warn(fmt.Sprintf(">> Unknown service: %v", service))
			break
		}
	}
	return m, nil
}
