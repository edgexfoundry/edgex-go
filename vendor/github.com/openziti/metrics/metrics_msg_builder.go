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
	"github.com/google/uuid"
	"github.com/openziti/metrics/metrics_pb"
	"github.com/rcrowley/go-metrics"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type messageBuilder metrics_pb.MetricsMessage

func newMessageBuilder(sourceId string, tags map[string]string) *messageBuilder {
	builder := &messageBuilder{
		EventId:   uuid.NewString(),
		Timestamp: timestamppb.New(time.Now()),
		SourceId:  sourceId,
		Tags:      tags,
	}

	return builder
}

func (builder *messageBuilder) addIntValue(name string, value int64) {
	if builder.IntValues == nil {
		builder.IntValues = make(map[string]int64)
	}
	builder.IntValues[name] = value
}

func (builder *messageBuilder) AddCounter(name string, metric metrics.Counter) {
	builder.addIntValue(name, metric.Count())
}

func (builder *messageBuilder) addIntGauge(name string, metric metrics.Gauge) {
	builder.addIntValue(name, metric.Value())
}

func (builder *messageBuilder) AddFloatGauge(name string, metric metrics.GaugeFloat64) {
	if builder.FloatValues == nil {
		builder.FloatValues = make(map[string]float64)
	}
	builder.FloatValues[name] = metric.Value()
}

func (builder *messageBuilder) addMeter(name string, metric metrics.Meter) {
	meter := &metrics_pb.MetricsMessage_Meter{}
	meter.Count = metric.Count()
	meter.M1Rate = metric.Rate1()
	meter.M5Rate = metric.Rate5()
	meter.M15Rate = metric.Rate15()
	meter.MeanRate = metric.RateMean()

	if builder.Meters == nil {
		builder.Meters = make(map[string]*metrics_pb.MetricsMessage_Meter)
	}

	builder.Meters[name] = meter
}

func (builder *messageBuilder) addHistogram(name string, metric metrics.Histogram) {
	histogram := &metrics_pb.MetricsMessage_Histogram{}
	histogram.Count = metric.Count()
	histogram.Max = metric.Max()
	histogram.Mean = metric.Mean()
	histogram.Min = metric.Min()
	histogram.StdDev = metric.StdDev()
	histogram.Variance = metric.Variance()

	ps := metric.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})

	histogram.P50 = ps[0]
	histogram.P75 = ps[1]
	histogram.P95 = ps[2]
	histogram.P99 = ps[3]
	histogram.P999 = ps[4]
	histogram.P9999 = ps[5]

	if builder.Histograms == nil {
		builder.Histograms = make(map[string]*metrics_pb.MetricsMessage_Histogram)
	}

	builder.Histograms[name] = histogram
}

func (builder *messageBuilder) addTimer(name string, metric metrics.Timer) {
	timer := &metrics_pb.MetricsMessage_Timer{}
	timer.Count = metric.Count()
	timer.Max = metric.Max()
	timer.Mean = metric.Mean()
	timer.Min = metric.Min()
	timer.StdDev = metric.StdDev()
	timer.Variance = metric.Variance()

	ps := metric.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})

	timer.P50 = ps[0]
	timer.P75 = ps[1]
	timer.P95 = ps[2]
	timer.P99 = ps[3]
	timer.P999 = ps[4]
	timer.P9999 = ps[5]

	timer.M1Rate = metric.Rate1()
	timer.M5Rate = metric.Rate5()
	timer.M15Rate = metric.Rate15()
	timer.MeanRate = metric.RateMean()

	if builder.Timers == nil {
		builder.Timers = make(map[string]*metrics_pb.MetricsMessage_Timer)
	}

	builder.Timers[name] = timer
}

func (builder *messageBuilder) addIntervalBucketEvents(events []*bucketEvent) {
	for _, event := range events {
		if builder.IntervalCounters == nil {
			builder.IntervalCounters = make(map[string]*metrics_pb.MetricsMessage_IntervalCounter)
		}
		counter, present := builder.IntervalCounters[event.name]
		if !present {
			builder.IntervalCounters[event.name] = event.interval
		} else {
			counter.Buckets = append(counter.Buckets, event.interval.Buckets...)
		}
	}
}
