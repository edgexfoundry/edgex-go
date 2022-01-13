//go:build !linux && !windows
// +build !linux,!windows

/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package telemetry

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

var LoggingClient logger.LoggingClient

func PollCpu() (cpuSnapshot CpuUsage) {
	if LoggingClient != nil {
		LoggingClient.Debug("could not poll CPU usage", "reason", "OS not compatible with metrics service")
	}

	return cpuSnapshot
}

func AvgCpuUsage(init, final CpuUsage) (avg float64) {
	if LoggingClient != nil {
		LoggingClient.Debug("could not average CPU usage", "reason", "OS not compatible with metrics service")
	}

	return -1
}
