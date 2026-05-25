/*******************************************************************************
 * Copyright 2022 Intel Corp.
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
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"

	"github.com/hashicorp/go-multierror"
	gometrics "github.com/rcrowley/go-metrics"
)

const (
	serviceNameTagKey     = "service"
	counterCountName      = "counter-count"
	gaugeValueName        = "gauge-value"
	gaugeFloat64ValueName = "gaugeFloat64-value"
	timerCountName        = "timer-count"
	timerMeanName         = "timer-mean"
	timerMinName          = "timer-min"
	timerMaxName          = "timer-max"
	timerStddevName       = "timer-stddev"
	timerVarianceName     = "timer-variance"
	histogramCountName    = "histogram-count"
	histogramMeanName     = "histogram-mean"
	histogramMinName      = "histogram-min"
	histogramMaxName      = "histogram-max"
	histogramStddevName   = "histogram-stddev"
	histogramVarianceName = "histogram-variance"
)

type messageBusReporter struct {
	lc               logger.LoggingClient
	serviceName      string
	dic              *di.Container
	messageClient    messaging.MessageClient
	config           *config.TelemetryInfo
	baseMetricsTopic string
}

// NewMessageBusReporter creates a new MessageBus reporter which reports metrics to the EdgeX MessageBus
func NewMessageBusReporter(lc logger.LoggingClient, baseTopic string, serviceName string, dic *di.Container, config *config.TelemetryInfo) interfaces.MetricsReporter {
	reporter := &messageBusReporter{
		lc:               lc,
		serviceName:      serviceName,
		dic:              dic,
		config:           config,
		baseMetricsTopic: common.BuildTopic(baseTopic, common.MetricsPublishTopic, serviceName),
	}

	return reporter
}

// Report collects all the current metrics and reports them to the EdgeX MessageBus
// The approach here was adapted from https://github.com/vrischmann/go-metrics-influxdb
func (r *messageBusReporter) Report(registry gometrics.Registry, metricTags map[string]map[string]string) error {
	var errs error
	publishedCount := 0

	// App Services create the messaging client after bootstrapping, so must get it from DIC when the first time
	if r.messageClient == nil {
		r.messageClient = container.MessagingClientFrom(r.dic.Get)
	}

	// If messaging client nil, then service hasn't set it up and can not report metrics this pass.
	// This may happen during bootstrapping if interval time is lower than time to bootstrap,
	// but will be resolved one messaging client has been added to the DIC.
	if r.messageClient == nil {
		return errors.New("messaging client not available. Unable to report metrics")
	}

	// Build the service tags each time we report since that can be changed in the Writable config
	serviceTags := buildMetricTags(r.config.Tags)
	serviceTags = append(serviceTags, dtos.MetricTag{
		Name:  serviceNameTagKey,
		Value: r.serviceName,
	})

	registry.Each(func(itemName string, item interface{}) {
		var nextMetric dtos.Metric
		var err error

		// If itemName matches a configured Metric name, use the configured Metric name in case it is a partial match.
		// The metric item will have the extra name portion as a tag.
		// This is important for Metrics for App Service Pipelines, when the Metric name reported need to be the same
		// for all pipelines, but each will have to have unique name (with pipeline ID added) registered.
		// The Pipeline id will also be added as a tag.
		name, isEnabled := r.config.GetEnabledMetricName(itemName)
		if !isEnabled {
			// This metric is not enable so do not report it.
			return
		}

		tags := append(serviceTags, buildMetricTags(metricTags[itemName])...)

		switch metric := item.(type) {
		case gometrics.Counter:
			snapshot := metric.Snapshot()
			fields := []dtos.MetricField{{Name: counterCountName, Value: snapshot.Count()}}
			nextMetric, err = dtos.NewMetric(name, fields, tags)

		case gometrics.Gauge:
			snapshot := metric.Snapshot()
			fields := []dtos.MetricField{{Name: gaugeValueName, Value: snapshot.Value()}}
			nextMetric, err = dtos.NewMetric(name, fields, tags)

		case gometrics.GaugeFloat64:
			snapshot := metric.Snapshot()
			fields := []dtos.MetricField{{Name: gaugeFloat64ValueName, Value: snapshot.Value()}}
			nextMetric, err = dtos.NewMetric(name, fields, tags)

		case gometrics.Timer:
			snapshot := metric.Snapshot()
			fields := []dtos.MetricField{
				{Name: timerCountName, Value: snapshot.Count()},
				{Name: timerMinName, Value: snapshot.Min()},
				{Name: timerMaxName, Value: snapshot.Max()},
				{Name: timerMeanName, Value: snapshot.Mean()},
				{Name: timerStddevName, Value: snapshot.StdDev()},
				{Name: timerVarianceName, Value: snapshot.Variance()},
			}
			nextMetric, err = dtos.NewMetric(name, fields, tags)

		case gometrics.Histogram:
			snapshot := metric.Snapshot()
			fields := []dtos.MetricField{
				{Name: histogramCountName, Value: snapshot.Count()},
				{Name: histogramMinName, Value: snapshot.Min()},
				{Name: histogramMaxName, Value: snapshot.Max()},
				{Name: histogramMeanName, Value: snapshot.Mean()},
				{Name: histogramStddevName, Value: snapshot.StdDev()},
				{Name: histogramVarianceName, Value: snapshot.Variance()},
			}
			nextMetric, err = dtos.NewMetric(name, fields, tags)

		default:
			errs = multierror.Append(errs, fmt.Errorf("metric type %T not supported", metric))
			return
		}

		if err != nil {
			err = fmt.Errorf("unable to create metric for '%s': %s", name, err.Error())
			errs = multierror.Append(errs, err)
			return
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, common.CorrelationHeader, uuid.NewString()) //nolint: staticcheck
		ctx = context.WithValue(ctx, common.ContentType, common.ContentTypeJSON) //nolint: staticcheck
		message := types.NewMessageEnvelope(nextMetric, ctx)

		topic := common.BuildTopic(r.baseMetricsTopic, name)
		if err := r.messageClient.Publish(message, topic); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to publish metric '%s' to topic '%s': %s", name, topic, err.Error()))
			return
		} else {
			publishedCount++
		}
	})

	r.lc.Debugf("Publish %d metrics to the '%s' base topic", publishedCount, r.baseMetricsTopic)

	return errs
}

func buildMetricTags(tags map[string]string) []dtos.MetricTag {
	var metricTags []dtos.MetricTag

	for tagName, tagValue := range tags {
		metricTags = append(metricTags, dtos.MetricTag{
			Name:  tagName,
			Value: tagValue,
		})
	}

	return metricTags
}
