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
)

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
)

func invokeAction(action string, services []string, params map[string]string) bool {

	// Loop through requested action, along with respectively-supplied parameters.
	for _, service := range services {
		// TODO: Logging is placeholder for actual functionality to be implemented.
		LoggingClient.Info(fmt.Sprintf("OPERATION >> About to {%v} the service {%v} with parameters {%v}.\n", action, service, params))
	}
	return true
}

func getConfig(services []string) bool {

	// Loop through requested actions, along with respectively-supplied parameters.
	for _, service := range services {
		// TODO: Logging is placeholder for actual functionality to be implemented.
		LoggingClient.Info(fmt.Sprintf("CONFIG >> Fetching the configuration for the service {%v}!\n", service))
	}
	return true
}

func getMetric(services []string, metrics []string) bool {

	// Loop through requested service(s), along with respectively-sought metric(s).
	for _, service := range services {
		for _, metric := range metrics {
			// TODO: Logging is placeholder for actual functionality to be implemented.
			// TODO: What constitutes metrics. How do services provide those metrics?
			LoggingClient.Info(fmt.Sprintf("METRIC >> Fetching the metric {%v} for the service {%v}...\n", metric, service))
		}
	}
	return true
}
