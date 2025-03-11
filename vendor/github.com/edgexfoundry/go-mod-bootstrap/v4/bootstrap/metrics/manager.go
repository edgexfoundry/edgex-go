/*******************************************************************************
 * Copyright 2022-2023 Intel Corp.
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

import (
	"context"
	"sync"
	"time"

	gometrics "github.com/rcrowley/go-metrics"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
)

type manager struct {
	lc         logger.LoggingClient
	metricTags map[string]map[string]string
	tagsMutex  *sync.RWMutex
	registry   gometrics.Registry
	reporter   interfaces.MetricsReporter
	interval   time.Duration
	ticker     *time.Ticker
}

func (m *manager) ResetInterval(interval time.Duration) {
	m.interval = interval
	if m.ticker == nil {
		return
	}

	m.ticker.Reset(m.interval)
	m.lc.Infof("Metrics Manager report interval changed to %s", m.interval.String())
}

// NewManager creates a new metrics manager
func NewManager(lc logger.LoggingClient, interval time.Duration, reporter interfaces.MetricsReporter) interfaces.MetricsManager {
	m := &manager{
		lc:         lc,
		registry:   gometrics.NewRegistry(),
		reporter:   reporter,
		interval:   interval,
		metricTags: make(map[string]map[string]string),
		tagsMutex:  new(sync.RWMutex),
	}

	return m
}

// Register registers a go-metric metric item which must be one of the
func (m *manager) Register(name string, item interface{}, tags map[string]string) error {
	if err := dtos.ValidateMetricName(name, "metric"); err != nil {
		return err
	}

	if len(tags) > 0 {
		if err := m.setMetricTags(name, tags); err != nil {
			return err
		}
	}

	if err := m.registry.Register(name, item); err != nil {
		return err
	}

	return nil
}

// IsRegistered checks whether a metric has been registered
func (m *manager) IsRegistered(name string) bool {
	return m.registry.Get(name) != nil
}

// Unregister unregisters a metric item
func (m *manager) Unregister(name string) {
	m.tagsMutex.Lock()
	defer m.tagsMutex.Unlock()

	m.registry.Unregister(name)
	m.metricTags[name] = nil
}

// Run periodically (based on configured interval) reports the collected metrics using the configured MetricsReporter.
func (m *manager) Run(ctx context.Context, wg *sync.WaitGroup) {

	m.ticker = time.NewTicker(m.interval)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				m.lc.Info("Exited Metrics Manager Run...")
				return

			case <-m.ticker.C:
				m.tagsMutex.RLock()
				tags := copyTagMaps(m.metricTags)
				m.tagsMutex.RUnlock()

				if err := m.reporter.Report(m.registry, tags); err != nil {
					m.lc.Errorf(err.Error())
					continue
				}

				m.lc.Debug("Reported metrics...")
			}
		}
	}()

	m.lc.Infof("Metrics Manager started with a report interval of %s", m.interval.String())
}

func copyTagMaps(origTagMaps map[string]map[string]string) map[string]map[string]string {
	tags := make(map[string]map[string]string)
	for key, value := range origTagMaps {
		tags[key] = copyTags(value)
	}

	return tags
}

func copyTags(origTags map[string]string) map[string]string {
	tags := make(map[string]string)
	for key, value := range origTags {
		tags[key] = value
	}

	return tags
}

// GetCounter retrieves the specified registered Counter
// Returns nil if named item not registered or not a Counter
func (m *manager) GetCounter(name string) gometrics.Counter {
	metric := m.registry.Get(name)
	if metric == nil {
		return nil
	}

	counter, ok := metric.(gometrics.Counter)
	if !ok {
		m.lc.Warnf("Unable to get Counter metric by name '%s': Registered metric by that name is not a Counter", name)
		return nil
	}

	return counter
}

// GetGauge retrieves the specified registered Gauge
// Returns nil if named item not registered or not a Gauge
func (m *manager) GetGauge(name string) gometrics.Gauge {
	metric := m.registry.Get(name)
	if metric == nil {
		return nil
	}

	gauge, ok := metric.(gometrics.Gauge)
	if !ok {
		m.lc.Warnf("Unable to get Gauge metric by name '%s': Registered metric by that name is not a Gauge", name)
		return nil
	}

	return gauge
}

// GetGaugeFloat64 retrieves the specified registered GaugeFloat64
// Returns nil if named item not registered or not a GaugeFloat64
func (m *manager) GetGaugeFloat64(name string) gometrics.GaugeFloat64 {
	metric := m.registry.Get(name)
	if metric == nil {
		return nil
	}

	gaugeFloat64, ok := metric.(gometrics.GaugeFloat64)
	if !ok {
		m.lc.Warnf("Unable to get GaugeFloat64 metric by name '%s': Registered metric by that name is not a GaugeFloat64", name)
		return nil
	}

	return gaugeFloat64
}

// GetTimer retrieves the specified registered Timer
// Returns nil if named item not registered or not a Timer
func (m *manager) GetTimer(name string) gometrics.Timer {
	metric := m.registry.Get(name)
	if metric == nil {
		return nil
	}

	timer, ok := metric.(gometrics.Timer)
	if !ok {
		m.lc.Warnf("Unable to get Timer metric by name '%s': Registered metric by that name is not a Timer", name)
		return nil
	}

	return timer
}

func (m *manager) setMetricTags(metricName string, tags map[string]string) error {
	for tagName := range tags {
		if err := dtos.ValidateMetricName(tagName, "Tag"); err != nil {
			return err
		}
	}

	m.tagsMutex.Lock()
	defer m.tagsMutex.Unlock()

	m.metricTags[metricName] = tags
	return nil
}
