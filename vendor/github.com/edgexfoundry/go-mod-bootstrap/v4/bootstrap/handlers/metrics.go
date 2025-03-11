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

package handlers

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/metrics"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

type RegisterTelemetryFunc func(logger.LoggingClient, *config.TelemetryInfo, interfaces.MetricsManager)

type ServiceMetrics struct {
	serviceName string
}

func NewServiceMetrics(serviceName string) *ServiceMetrics {
	return &ServiceMetrics{
		serviceName: serviceName,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization of service metrics.
func (s *ServiceMetrics) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := container.LoggingClientFrom(dic.Get)
	serviceConfig := container.ConfigurationFrom(dic.Get)

	telemetryConfig := serviceConfig.GetTelemetryInfo()

	if telemetryConfig.Interval == "" {
		telemetryConfig.Interval = "0s"
	}

	interval, err := time.ParseDuration(telemetryConfig.Interval)
	if err != nil {
		lc.Errorf("Telemetry interval is invalid time duration: %s", err.Error())
		return false
	}

	if interval == 0 {
		lc.Infof("0 specified for metrics reporting interval. Setting to max duration to effectively disable reporting.")
		interval = math.MaxInt64
	}

	baseTopic := serviceConfig.GetBootstrap().MessageBus.GetBaseTopicPrefix()
	reporter := metrics.NewMessageBusReporter(lc, baseTopic, s.serviceName, dic, telemetryConfig)
	manager := metrics.NewManager(lc, interval, reporter)

	manager.Run(ctx, wg)

	dic.Update(di.ServiceConstructorMap{
		container.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return manager
		},
	})

	return true
}
