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
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
)

type SystemUsage struct {
	Memory     memoryUsage
	CpuBusyAvg float64
}

type memoryUsage struct {
	Alloc,
	TotalAlloc,
	Sys,
	Mallocs,
	Frees,
	LiveObjects uint64
}

type CpuUsage struct {
	Busy, // time used by all processes. this ideally does not include system processes.
	Idle, // time used by the idle process
	Total uint64 // reported sum total of all usage
}

var once sync.Once
var lastSample CpuUsage
var usageAvg float64
var wg sync.WaitGroup

func NewSystemUsage() (s SystemUsage) {
	// The micro-service is to be considered the System Of Record (SOR) in terms of accurate information.
	// Fetch metrics for the metadata service.
	var rtm runtime.MemStats

	// Read full memory stats
	runtime.ReadMemStats(&rtm)

	// Miscellaneous memory stats
	s.Memory.Alloc = rtm.Alloc
	s.Memory.TotalAlloc = rtm.TotalAlloc
	s.Memory.Sys = rtm.Sys
	s.Memory.Mallocs = rtm.Mallocs
	s.Memory.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	s.Memory.LiveObjects = s.Memory.Mallocs - s.Memory.Frees

	s.CpuBusyAvg = usageAvg

	return s
}

// deprecated
func StartCpuUsageAverage() {
	once.Do(func() {
		for {
			wg.Add(1)
			nextUsage := PollCpu()
			usageAvg = AvgCpuUsage(lastSample, nextUsage)
			lastSample = nextUsage
			wg.Done()

			time.Sleep(time.Second * 30)
		}
	})
}

// BootstrapHandler fulfills the BootstrapHandler contract.  It creates a go routine to periodically sample CPU usage
// and is intended to supersede the existing StartCpuUsageAverage() function when the new bootstrap package is used
// by all of the core services.
func BootstrapHandler(
	wg *sync.WaitGroup,
	ctx context.Context,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	loggingClient := container.LoggingClientFrom(dic.Get)
	loggingClient.Info("Telemetry starting")

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			nextUsage := PollCpu()
			usageAvg = AvgCpuUsage(lastSample, nextUsage)
			lastSample = nextUsage

			for seconds := 30; seconds > 0; seconds-- {
				select {
				case <-ctx.Done():
					loggingClient.Info("Telemetry stopped")
					return
				default:
					time.Sleep(time.Second)
				}
			}
		}
	}()

	return true
}
