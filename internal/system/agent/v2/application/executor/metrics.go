//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"context"
	"net/http"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/system"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/response"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
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
func (m *metrics) Get(_ context.Context, services []string) (interface{}, errors.EdgeX) {
	metrics := make(map[string]interface{})

	var wg sync.WaitGroup
	var errCh = make(chan error, len(services))
	for _, service := range services {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()

			raw, err := m.executor(m.executorPath, serviceName, system.Metrics)
			if err != nil {
				errCh <- err
				return
			}

			r := response.Process(raw, m.lc)
			metrics[serviceName] = r
		}(service)
	}

	wg.Wait()
	close(errCh)

	err := <-errCh
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindUnknown, "", err)
	}

	res := common.NewMultiMetricsResponse("", "", http.StatusOK, metrics)
	return res, nil
}
