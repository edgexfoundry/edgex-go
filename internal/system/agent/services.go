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
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/executor"
	"time"
)

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
)

func InvokeOperation(action string, services []string) bool {

	// Loop through requested operation, along with respectively-supplied parameters.
	for _, service := range services {
		LoggingClient.Info(fmt.Sprintf("About to {%v} the service {%v} ", action, service))

		if !isKnownServiceKey(service) {
			LoggingClient.Warn(fmt.Sprintf("unknown service: %v", service))
		}

		switch action {
		case START:
			executor.StartDockerContainerCompose(service, Configuration.ComposeUrl)
			break

		case STOP:
			ec.StopService(service)
			break

		case RESTART:
			ec.StopService(service)
			time.Sleep(time.Second * time.Duration(1))
			executor.StartDockerContainerCompose(service, Configuration.ComposeUrl)
			break
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

		if !isKnownServiceKey(service) {
			LoggingClient.Warn(fmt.Sprintf("unknown service: %v", service))
		}

		responseJSON, err := clients[service].FetchConfiguration()
		if err != nil {
			msg := fmt.Sprintf("%s get config error: %s", service, err.Error())
			c.Configuration[service] = msg
			LoggingClient.Error(msg)
		} else {
			c.Configuration[service] = ProcessResponse(responseJSON)
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

		if !isKnownServiceKey(service) {
			LoggingClient.Warn(fmt.Sprintf("unknown service: %v", service))
		}

		responseJSON, err := clients[service].FetchMetrics()
		if err != nil {
			msg := fmt.Sprintf("%s get metrics error: %s", service, err.Error())
			m.Metrics[service] = msg
			LoggingClient.Error(msg)
		} else {
			m.Metrics[service] = ProcessResponse(responseJSON)
		}
	}
	return m, nil
}

func isKnownServiceKey(serviceKey string) bool {
	// create a map because this is the easiest/cleanest way to determine whether something exists in a set
	var services = map[string]struct{}{
		internal.SupportNotificationsServiceKey: {},
		internal.CoreCommandServiceKey: {},
		internal.CoreDataServiceKey: {},
		internal.CoreMetaDataServiceKey: {},
		internal.ExportClientServiceKey: {},
		internal.ExportDistroServiceKey: {},
		internal.SupportLoggingServiceKey: {},
		internal.SupportSchedulerServiceKey: {},
		internal.ConfigSeedServiceKey: {},
	}

	 _, exists := services[serviceKey]

	 return exists
}
