/*******************************************************************************
 * Copyright 2019 VMware Inc.
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

import "github.com/edgexfoundry/edgex-go/internal/system/executor"

const (
	operation     = "operation"
	configuration = "config"
	services      = "services"
	start         = executor.Start
	stop          = executor.Stop
	restart       = executor.Restart
	metrics       = executor.Metrics
	health        = "health"
	ping          = "ping"

	metricsOptionViaDirectService = "direct-service"
	metricsOptionViaExecutor      = "executor"

	executorTypeDirectService = metricsOptionViaDirectService
	executorTypeUnknown       = "unknown"
)
