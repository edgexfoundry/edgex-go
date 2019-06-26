/*******************************************************************************
 * Copyright 2019 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	servicesMock "github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/mock"
)

/* 	NOTE: The following two "services-related" functions are tested _elsewhere_:
	getConfig(...)
	getMetrics(...)
Those (two "services-related" functions ) are tested in "go-mod-core-contracts", specifically in:
	/github.com/edgexfoundry/go-mod-core-contracts/clients/general/client_test.go
Therefore, coverage for those two is not provided _here_.
*/

func reset() {
	Configuration = &ConfigurationStruct{}
}

func TestStartOperation(t *testing.T) {
	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockStarter := &servicesMock.ServiceStarter{}
	mockStarter.On("Start", mock.AnythingOfType("string")).Return(nil)

	tests := []struct {
		name        string
		starter     interface{}
		services    []string
		expectError bool
	}{
		{"start services", mockStarter, serviceList, false},
		{"type check failure", "abc", serviceList, true},
	}
	for _, tt := range tests {
		executorClient = tt.starter
		t.Run(tt.name, func(t *testing.T) {
			err := InvokeOperation("start", tt.services)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err == nil {
				t.Errorf("did not receive expected error: %s", tt.name)
			}
		})
	}
}

func TestStopOperation(t *testing.T) {
	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockStopper := &servicesMock.ServiceStopper{}
	mockStopper.On("Stop", mock.AnythingOfType("string")).Return(nil)

	tests := []struct {
		name        string
		stopper     interface{}
		services    []string
		expectError bool
	}{
		{"stop services", mockStopper, serviceList, false},
		{"type check failure", "xyz", serviceList, true},
	}
	for _, tt := range tests {
		executorClient = tt.stopper
		t.Run(tt.name, func(t *testing.T) {
			err := InvokeOperation("stop", tt.services)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err == nil {
				t.Errorf("did not receive expected error: %s", tt.name)
			}
		})
	}
}

func TestRestartOperation(t *testing.T) {
	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockRestarter := &servicesMock.ServiceRestarter{}
	mockRestarter.On("Restart", mock.AnythingOfType("string")).Return(nil)

	tests := []struct {
		name        string
		restarter   interface{}
		services    []string
		expectError bool
	}{
		{"restart services", mockRestarter, serviceList, false},
		{"type check failure", "qrs", serviceList, true},
	}
	for _, tt := range tests {
		executorClient = tt.restarter
		t.Run(tt.name, func(t *testing.T) {
			err := InvokeOperation("restart", tt.services)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err == nil {
				t.Errorf("did not receive expected error: %s", tt.name)
			}
		})
	}
}

func TestUnsupportedService(t *testing.T) {
	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{"unknown service"}
	mockStarter := &servicesMock.ServiceStarter{}
	mockStarter.On("Start", mock.AnythingOfType("string")).Return(nil)

	tests := []struct {
		name        string
		starter     interface{}
		services    []string
		expectError bool
	}{
		{"check failure when service is unknown", mockStarter, serviceList, true},
	}
	for _, tt := range tests {
		executorClient = tt.starter
		t.Run(tt.name, func(t *testing.T) {
			err := InvokeOperation("start", tt.services)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestInvokeMetricsWithExecutor(t *testing.T) {

	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockMetricsFetcher := &servicesMock.MetricsFetcher{}

	mockInvokeMetricsResultFn := func(services []string, ctx context.Context) MetricsRespMap {
		metricsRespMap := MetricsRespMap{Metrics: ProcessResponse(METRICS)}

		return metricsRespMap
	}

	mockMetricsFetcher.On("Metrics", context.Background(), mock.AnythingOfType("string")).Return(mockInvokeMetricsResultFn)

	// Create a new context.Context and populate it with data.
	ctx := context.Background()
	ctx = context.WithValue(ctx, "app.auth.token", "abc123")

	tests := []struct {
		name        string
		fetcher     interface{}
		services    []string
		expectError bool
		config      string
	}{
		{"invoke metrics", mockMetricsFetcher, serviceList, false, "executor"},
		{"type check failure", "abc", serviceList, true, "executor"},
	}
	for _, tt := range tests {
		executorClient = tt.fetcher
		t.Run(tt.name, func(t *testing.T) {
			Configuration.MetricsMechanism = tt.config
			m, err := InvokeMetrics(tt.services, ctx)
			if !tt.expectError && err != nil && m.Metrics == nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err == nil && m.Metrics != nil {
				t.Errorf("did not receive expected error: %s", tt.name)
			}
		})
	}
}

func TestInvokeMetricsWithCustom(t *testing.T) {

	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockMetricsFetcher := &servicesMock.MetricsFetcher{}

	mockInvokeMetricsResultFn := func(services []string, ctx context.Context) MetricsRespMap {
		metricsRespMap := MetricsRespMap{Metrics: ProcessResponse(METRICS)}

		return metricsRespMap
	}

	mockMetricsFetcher.On("Metrics", context.Background(), mock.AnythingOfType("string")).Return(mockInvokeMetricsResultFn)

	// Create a new context.Context and populate it with data.
	ctx := context.Background()
	ctx = context.WithValue(ctx, "app.auth.token", "abc123")

	tests := []struct {
		name        string
		fetcher     interface{}
		services    []string
		expectError bool
		config      string
	}{
		{"invoke metrics", mockMetricsFetcher, serviceList, true, "custom"},
		{"type check failure", "abc", serviceList, true, "custom"},
	}
	for _, tt := range tests {
		executorClient = tt.fetcher
		t.Run(tt.name, func(t *testing.T) {
			Configuration.MetricsMechanism = tt.config
			_, err := InvokeMetrics(tt.services, ctx)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestInvokeMetricsWithUnsupported(t *testing.T) {

	reset()
	LoggingClient = logger.NewMockClient()
	serviceList := []string{clients.CoreDataServiceKey, clients.CoreCommandServiceKey}
	mockMetricsFetcher := &servicesMock.MetricsFetcher{}

	mockInvokeMetricsResultFn := func(services []string, ctx context.Context) MetricsRespMap {
		metricsRespMap := MetricsRespMap{Metrics: ProcessResponse(METRICS)}

		return metricsRespMap
	}

	mockMetricsFetcher.On("Metrics", context.Background(), mock.AnythingOfType("string")).Return(mockInvokeMetricsResultFn)

	// Create a new context.Context and populate it with data.
	ctx := context.Background()
	ctx = context.WithValue(ctx, "app.auth.token", "abc123")

	tests := []struct {
		name        string
		fetcher     interface{}
		services    []string
		expectError bool
		config      string
	}{
		{"invoke metrics", mockMetricsFetcher, serviceList, true, "unsupported"},
		{"type check failure", "abc", serviceList, true, "unsupported"},
	}
	for _, tt := range tests {
		executorClient = tt.fetcher
		t.Run(tt.name, func(t *testing.T) {
			Configuration.MetricsMechanism = tt.config
			_, err := InvokeMetrics(tt.services, ctx)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestProcessOutput(t *testing.T) {

	// Setup
	stdoutDataGood := []byte("level=INFO ts=2019-08-15T15:56:08.959259Z app=docker-compose-executor source=main.go:58 msg=\"This is the docker-compose-executor application!\"\r\nlevel=INFO ts=2019-08-15T15:56:08.959572Z app=docker-compose-executor source=main.go:61 msg=\"Application started in: 322.465\u00B5s\"\r\nlevel=INFO ts=2019-08-15T15:56:11.159011Z app=docker-compose-executor source=main.go:75 msg=\"success performing metrics on service edgex-core-command\"\r\n{\"cpu_perc\":\"0.00%\",\"mem_usage\":\"4.766MiB / 1.952GiB\",\"mem_perc\":\"0.24%\",\"net_io\":\"8.49MB / 4.71MB\",\"block_io\":\"10.2MB / 0B\",\"pids\":\"18\"}\r\n")
	expected := map[string]interface{}{"cpu_perc": "0.00%", "mem_usage": "4.766MiB / 1.952GiB", "mem_perc": "0.24%", "net_io": "8.49MB / 4.71MB", "block_io": "10.2MB / 0B", "pids": "18"}

	// Invoke business logic and validate results
	result, _ := processOutput(stdoutDataGood)
	s := string(result)
	assert.NotNil(t, s)

	actual := make(map[string]interface{})
	actual[clients.CoreDataServiceKey] = ProcessResponse(s)
	observed := actual[clients.CoreDataServiceKey]
	if !reflect.DeepEqual(expected, observed) {
		t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", expected, actual)
		return
	}

	// Setup
	stdoutDataBad := []byte("")
	exp := ""

	// Invoke business logic and validate results
	result, _ = processOutput(stdoutDataBad)
	s = string(result)
	if !assert.Equal(t, exp, s) {
		t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", exp, s)
		return
	}
}
