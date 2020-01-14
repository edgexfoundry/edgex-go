/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package metrics

import "github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

// Service contains references to dependencies required by the corresponding implementation.
type Service struct{}

// NewService is a factory function that returns an initialized Service receiver struct.
func NewService() *Service {
	return &Service{}
}

// Get gathers metrics from telemetry's implementation and returns them.
func (s *Service) Get() (alloc, totalAlloc, sys, mallocs, frees, liveObjects uint64, cpuBusyAvg float64) {
	m := telemetry.NewSystemUsage()
	return m.Memory.Alloc,
		m.Memory.TotalAlloc,
		m.Memory.Sys,
		m.Memory.Mallocs,
		m.Memory.Frees,
		m.Memory.LiveObjects,
		m.CpuBusyAvg
}
