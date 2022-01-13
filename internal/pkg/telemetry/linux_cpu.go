//go:build linux
// +build linux

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
	"math"

	"bitbucket.org/bertimus9/systemstat"
)

func PollCpu() (cpuSnapshot CpuUsage) {
	linuxSample := systemstat.GetCPUSample()
	return CpuUsage{
		Busy:  linuxSample.Nice + linuxSample.User,
		Idle:  linuxSample.Idle,
		Total: linuxSample.Total,
	}
}

func AvgCpuUsage(init, final CpuUsage) (avg float64) {
	// SimpleAverage only uses idle and total, so only copy those
	linuxInit := systemstat.CPUSample{
		Idle:  init.Idle,
		Total: init.Total,
	}

	linuxFinal := systemstat.CPUSample{
		Idle:  final.Idle,
		Total: final.Total,
	}

	avg = systemstat.GetSimpleCPUAverage(linuxInit, linuxFinal).BusyPct

	if avg < .000001 || math.IsNaN(avg) {
		return 0.0
	}

	return avg
}
