/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package metrics

import (
	"fmt"
	"reflect"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/metrics/metrics_pb"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/rcrowley/go-metrics"
)

// Metric is the base functionality for all metrics types
type Metric interface {
	Dispose()
}

// Registry allows for configuring and accessing metrics for an application
type Registry interface {
	// SourceId returns the source id of this Registry
	SourceId() string

	// Gauge returns a Gauge for the given name. If one does not yet exist, one will be created
	Gauge(name string) Gauge

	// FuncGauge returns a Gauge for the given name. If one does not yet exist, one will be created using
	// the given function
	FuncGauge(name string, f func() int64) Gauge

	// Meter returns a Meter for the given name. If one does not yet exist, one will be created
	Meter(name string) Meter

	// Histogram returns a Histogram for the given name. If one does not yet exist, one will be created
	Histogram(name string) Histogram

	// Timer returns a Timer for the given name. If one does not yet exist, one will be created
	Timer(name string) Timer

	// EachMetric calls the given visitor function for each Metric in this registry
	EachMetric(visitor func(name string, metric Metric))

	// GetGauge returns the Gauge for the given name or nil if a Gauge with that name doesn't exist
	GetGauge(name string) Gauge

	// GetMeter returns the Meter for the given name or nil if a Meter with that name doesn't exist
	GetMeter(name string) Meter

	// GetHistogram returns the Histogram for the given name or nil if a Histogram with that name doesn't exist
	GetHistogram(name string) Histogram

	// GetTimer returns the Timer for the given name or nil if a Timer with that name doesn't exist
	GetTimer(name string) Timer

	// IsValidMetric returns true if a metric with the given name exists in the registry, false otherwise
	IsValidMetric(name string) bool

	// Poll returns a MetricsMessage with a snapshot of the metrics in the Registry
	Poll() *metrics_pb.MetricsMessage

	// DisposeAll removes and cleans up all metrics currently in the Registry
	DisposeAll()
}

func NewRegistry(sourceId string, tags map[string]string) Registry {
	return &registryImpl{
		sourceId:  sourceId,
		tags:      tags,
		metricMap: cmap.New[any](),
	}
}

type registryImpl struct {
	sourceId  string
	tags      map[string]string
	metricMap cmap.ConcurrentMap[string, any]
}

func (registry *registryImpl) dispose(name string) {
	registry.metricMap.Remove(name)
}

func (registry *registryImpl) DisposeAll() {
	registry.EachMetric(func(name string, metric Metric) {
		metric.Dispose()
	})
}

func (registry *registryImpl) IsValidMetric(name string) bool {
	return registry.metricMap.Has(name)
}

func (registry *registryImpl) SourceId() string {
	return registry.sourceId
}

func (registry *registryImpl) GetGauge(name string) Gauge {
	metric, found := registry.metricMap.Get(name)
	if !found {
		return nil
	}
	if gauge, ok := metric.(Gauge); ok {
		return gauge
	}
	return nil
}

func (registry *registryImpl) GetMeter(name string) Meter {
	metric, found := registry.metricMap.Get(name)
	if !found {
		return nil
	}
	if meter, ok := metric.(Meter); ok {
		return meter
	}
	return nil
}

func (registry *registryImpl) GetHistogram(name string) Histogram {
	metric, found := registry.metricMap.Get(name)
	if !found {
		return nil
	}
	if histogram, ok := metric.(Histogram); ok {
		return histogram
	}
	return nil
}

func (registry *registryImpl) GetTimer(name string) Timer {
	metric, found := registry.metricMap.Get(name)
	if !found {
		return nil
	}
	if timer, ok := metric.(Timer); ok {
		return timer
	}
	return nil
}

func getOrCreateMetric[T Metric](registry *registryImpl, name string, newMetric func() T) T {
	var result T
	for {
		metric, present := registry.metricMap.Get(name)
		if present {
			var ok bool
			result, ok = metric.(T)
			if !ok {
				panic(fmt.Errorf("metric '%v' already exists and is not a %T. It is a %T", name, new(T), metric))
			}
			return result
		}

		result = newMetric()
		if registry.metricMap.SetIfAbsent(name, result) {
			return result
		}
	}
}

func (registry *registryImpl) Gauge(name string) Gauge {
	return getOrCreateMetric(registry, name, func() Gauge {
		return &gaugeImpl{
			Gauge: metrics.NewGauge(),
			dispose: func() {
				registry.dispose(name)
			},
		}
	})
}

func (registry *registryImpl) FuncGauge(name string, f func() int64) Gauge {
	return getOrCreateMetric(registry, name, func() Gauge {
		return &gaugeImpl{
			Gauge: metrics.NewFunctionalGauge(f),
			dispose: func() {
				registry.dispose(name)
			},
		}
	})
}

func (registry *registryImpl) newMeter(name string) *meterImpl {
	return &meterImpl{
		Meter:    metrics.NewMeter(),
		registry: registry,
		name:     name,
	}
}

func (registry *registryImpl) Meter(name string) Meter {
	metric := registry.getRefCounted(name, func() refCounted {
		return registry.newMeter(name)
	})

	meter, ok := metric.(Meter)
	if !ok {
		panic(fmt.Errorf("metric '%v' already exists and is not a meter. It is a %v", name, reflect.TypeOf(metric).Name()))
	}
	return meter
}

func (registry *registryImpl) newHistogram(name string) *histogramImpl {
	return &histogramImpl{
		Histogram: metrics.NewHistogram(metrics.NewExpDecaySample(128, 0.015)),
		registry:  registry,
		name:      name,
	}
}

func (registry *registryImpl) Histogram(name string) Histogram {
	metric := registry.getRefCounted(name, func() refCounted {
		return registry.newHistogram(name)
	})

	histogram, ok := metric.(Histogram)
	if !ok {
		panic(fmt.Errorf("metric '%v' already exists and is not a histogram. It is a %v", name, reflect.TypeOf(metric).Name()))
	}
	return histogram
}

func (registry *registryImpl) getRefCounted(name string, factory func() refCounted) refCounted {
	metric := registry.metricMap.Upsert(name, nil, func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
		if exist {
			if h, ok := valueInMap.(refCounted); ok {
				h.IncrRefCount()
			}
			return valueInMap
		}

		newVal := factory()
		newVal.IncrRefCount()
		return newVal
	})

	histogram, ok := metric.(refCounted)
	if !ok {
		panic(fmt.Errorf("metric '%v' already exists and is not an instance of refCouted. It is a %v", name, reflect.TypeOf(metric).Name()))
	}
	return histogram
}

func (registry *registryImpl) disposeRefCounted(metric refCounted) {
	registry.metricMap.RemoveCb(metric.Name(), func(key string, v interface{}, exists bool) bool {
		if !exists {
			return true
		}
		return v == metric && metric.DecrRefCount() < 1
	})
}

func (registry *registryImpl) Timer(name string) Timer {
	return getOrCreateMetric(registry, name, func() Timer {
		return &timerImpl{
			Timer: metrics.NewTimer(),
			dispose: func() {
				registry.dispose(name)
			},
		}
	})
}

func (registry *registryImpl) EachMetric(visitor func(name string, metric Metric)) {
	for entry := range registry.metricMap.IterBuffered() {
		visitor(entry.Key, entry.Val.(Metric))
	}
}

func (registry *registryImpl) Each(visitor func(string, interface{})) {
	for entry := range registry.metricMap.IterBuffered() {
		visitor(entry.Key, entry.Val)
	}
}

// Provide rest of go-metrics Registry interface, so we can use go-metrics reporters if desired
func (registry *registryImpl) Get(s string) interface{} {
	val, _ := registry.metricMap.Get(s)
	return val
}

func (registry *registryImpl) GetAll() map[string]map[string]interface{} {
	return nil
}

func (registry *registryImpl) GetOrRegister(s string, i interface{}) interface{} {
	return registry.metricMap.Upsert(s, i, func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
		if exist {
			return valueInMap
		}
		return newValue
	})
}

func (registry *registryImpl) Register(s string, i interface{}) error {
	if registry.metricMap.SetIfAbsent(s, i) {
		return fmt.Errorf("duplicate metric %v", s)
	}
	return nil
}

func (registry *registryImpl) RunHealthchecks() {
}

func (registry *registryImpl) Unregister(s string) {
	registry.metricMap.Remove(s)
}

func (registry *registryImpl) UnregisterAll() {
	for _, key := range registry.metricMap.Keys() {
		registry.Unregister(key)
	}
}

func (registry *registryImpl) Poll() *metrics_pb.MetricsMessage {
	// If there's nothing to report, skip it
	if registry.metricMap.Count() == 0 {
		return nil
	}

	builder := newMessageBuilder(registry.sourceId, registry.tags)

	registry.EachMetric(func(name string, i Metric) {
		switch metric := i.(type) {
		case *gaugeImpl:
			builder.addIntGauge(name, metric.Snapshot())
		case *meterImpl:
			builder.addMeter(name, metric.Snapshot())
		case *histogramImpl:
			builder.addHistogram(name, metric.Snapshot())
		case *timerImpl:
			builder.addTimer(name, metric.Snapshot())
		case *intervalCounterImpl:
		// ignore, handled below
		case *usageCounterImpl:
			// ignore, handled below
		default:
			pfxlog.Logger().Errorf("Unsupported metric type %v", reflect.TypeOf(i))
		}
	})

	return (*metrics_pb.MetricsMessage)(builder)
}

type refCounted interface {
	IncrRefCount() int32
	DecrRefCount() int32
	Name() string
	stop()
}
