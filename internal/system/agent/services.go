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
	"context"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
)

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
)

func InvokeOperation(action string, services []string) bool {

	// Loop through requested operation, along with respectively-supplied parameters.
	for _, service := range services {
		logs.LoggingClient.Info("invoking operation on service", "action type", action, "service name", service)

		if !isKnownServiceKey(service) {
			logs.LoggingClient.Warn("unknown service found during invocation", "service name", service)
		}

		switch action {

		case START:
			if starter, ok := executorClient.(interfaces.ServiceStarter); ok {
				err := starter.Start(service)
				if err != nil {
					msg := fmt.Sprintf("error starting service \"%s\": %v", service, err)
					logs.LoggingClient.Error(msg)
				}
			} else {
				msg := fmt.Sprintf("starting not supported with specified executor")
				logs.LoggingClient.Warn(msg)
			}

		case STOP:
			if stopper, ok := executorClient.(interfaces.ServiceStopper); ok {
				err := stopper.Stop(service)
				if err != nil {
					msg := fmt.Sprintf("error stopping service \"%s\": %v", service, err)
					logs.LoggingClient.Error(msg)
				}
			} else {
				msg := fmt.Sprintf("stopping not supported with specified executor")
				logs.LoggingClient.Warn(msg)
			}

		case RESTART:
			if restarter, ok := executorClient.(interfaces.ServiceRestarter); ok {
				err := restarter.Restart(service)
				if err != nil {
					msg := fmt.Sprintf("error restarting service \"%s\": %v", service, err)
					logs.LoggingClient.Error(msg)
				}
			} else {
				msg := fmt.Sprintf("restarting not supported with specified executor")
				logs.LoggingClient.Warn(msg)
			}
		}
	}
	return true
}

func getConfig(services []string, ctx context.Context) (ConfigRespMap, error) {

	c := ConfigRespMap{}
	c.Configuration = map[string]interface{}{}

	// Loop through requested actions, along with respectively-supplied parameters.
	for _, service := range services {

		c.Configuration[service] = ""

		if !isKnownServiceKey(service) {
			logs.LoggingClient.Warn("unknown service found getting configuration", "service name", service)
		}

		responseJSON, err := clients[service].FetchConfiguration(ctx)
		if err != nil {
			c.Configuration[service] = fmt.Sprintf("%s get config error: %s", service, err.Error())
			logs.LoggingClient.Error("error retrieving configuration", "service name", service, "error message", err.Error())
		} else {
			c.Configuration[service] = ProcessResponse(responseJSON)
		}
	}
	return c, nil
}

func getMetrics(services []string, ctx context.Context) (MetricsRespMap, error) {

	m := MetricsRespMap{}
	m.Metrics = map[string]interface{}{}

	// Loop through requested actions, along with respectively-supplied parameters.
	for _, service := range services {

		m.Metrics[service] = ""

		if !isKnownServiceKey(service) {
			logs.LoggingClient.Warn("unknown service found getting metrics", "service name", service)
		}

		responseJSON, err := clients[service].FetchMetrics(ctx)
		if err != nil {
			m.Metrics[service] = fmt.Sprintf("%s get metrics error: %s", service, err.Error())
			logs.LoggingClient.Error("error retrieving metrics", "service name", service, "error message", err.Error())
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
		internal.CoreCommandServiceKey:          {},
		internal.CoreDataServiceKey:             {},
		internal.CoreMetaDataServiceKey:         {},
		internal.ExportClientServiceKey:         {},
		internal.ExportDistroServiceKey:         {},
		internal.SupportLoggingServiceKey:       {},
		internal.SupportSchedulerServiceKey:     {},
		internal.ConfigSeedServiceKey:           {},
	}

	_, exists := services[serviceKey]

	return exists
}
