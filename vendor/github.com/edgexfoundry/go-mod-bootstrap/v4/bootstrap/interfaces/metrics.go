//
// Copyright (c) 2022 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interfaces

import (
	"context"
	"sync"
	"time"

	gometrics "github.com/rcrowley/go-metrics"
)

// MetricsManager manages a services metrics
type MetricsManager interface {
	// ResetInterval resets the interval between reporting the current metrics
	ResetInterval(interval time.Duration)
	// Register registers a go-metrics metric item such as a Counter
	Register(name string, item interface{}, tags map[string]string) error
	// IsRegistered checks whether a metric has been registered
	IsRegistered(name string) bool
	// Unregister unregisters a go-metrics metric item such as a Counter
	Unregister(name string)
	// Run starts the collection of metrics
	Run(ctx context.Context, wg *sync.WaitGroup)
	// GetCounter retrieves the specified registered Counter
	// Returns nil if named item not registered or not a Counter
	GetCounter(name string) gometrics.Counter
	// GetGauge retrieves the specified registered Gauge
	// Returns nil if named item not registered or not a Gauge
	GetGauge(name string) gometrics.Gauge
	// GetGaugeFloat64 retrieves the specified registered GaugeFloat64
	// Returns nil if named item not registered or not a GaugeFloat64
	GetGaugeFloat64(name string) gometrics.GaugeFloat64
	// GetTimer retrieves the specified registered Timer
	// Returns nil if named item not registered or not a Timer
	GetTimer(name string) gometrics.Timer
}

// MetricsReporter reports the metrics
type MetricsReporter interface {
	Report(registry gometrics.Registry, metricTags map[string]map[string]string) error
}
