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
package scheduler

import "github.com/edgexfoundry/edgex-go/internal/pkg/config"

// Configuration V2 for the Support Scheduler Service
type ConfigurationStruct struct {

	ScheduleIntervalTime    int
	Clients         map[string]config.ClientInfo
	Databases       map[string]config.DatabaseInfo
	Logging         config.LoggingInfo
	Registry        config.RegistryInfo
	Service         config.ServiceInfo
	Intervals       map[string]config.IntervalInfo
	IntervalActions map[string]config.IntervalActionInfo
}
