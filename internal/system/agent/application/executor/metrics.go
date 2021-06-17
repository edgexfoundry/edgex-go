//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"context"
	"net/http"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/response"
)

// metrics contains references to dependencies required to handle the metrics via external executor use case.
type metrics struct {
	executor     interfaces.CommandExecutor
	lc           logger.LoggingClient
	executorPath string
}

// NewMetrics is a factory function that returns an initialized metrics receiver struct.
func NewMetrics(executor interfaces.CommandExecutor, lc logger.LoggingClient, executorPath string) *metrics {
	return &metrics{
		executor:     executor,
		lc:           lc,
		executorPath: executorPath,
	}
}

// Get implements the Metrics interface to obtain metrics via executor for one or more services concurrently.
func (m *metrics) Get(_ context.Context, services []string) ([]interface{}, errors.EdgeX) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var responses []interface{}

	for _, service := range services {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()

			res, err := m.executor(m.executorPath, edgexPrefix+serviceName, "metrics")
			if err != nil {
				mu.Lock()
				responses = append(responses, common.BaseWithMetricsResponse{
					BaseResponse: common.NewBaseResponse("", res, http.StatusInternalServerError),
					ServiceName:  serviceName,
					Metrics:      nil,
				})
				mu.Unlock()
				return
			}

			r := response.Process(res, m.lc)
			mu.Lock()
			responses = append(responses, common.BaseWithMetricsResponse{
				BaseResponse: common.NewBaseResponse("", "", http.StatusOK),
				ServiceName:  serviceName,
				Metrics:      r,
			})
			mu.Unlock()
		}(service)
	}

	wg.Wait()
	return responses, nil
}
