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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
)

// operation contains references to dependencies required to handle the operation via external executor use case.
type operation struct {
	executor     interfaces.CommandExecutor
	executorPath string
	lc           logger.LoggingClient
}

// NewOperation is a factory function that returns an initialized operation receiver struct.
func NewOperation(executor interfaces.CommandExecutor, executorPath string, lc logger.LoggingClient) *operation {
	return &operation{
		executor:     executor,
		executorPath: executorPath,
		lc:           lc,
	}
}

// Do concurrently delegates a start/stop/restart operation request to the configuration-defined executor.
func (o *operation) Do(_ context.Context, operations []requests.OperationRequest) ([]interface{}, errors.EdgeX) {
	var wg sync.WaitGroup
	var responses []interface{}

	for _, operation := range operations {
		wg.Add(1)
		go func(operation requests.OperationRequest) {
			defer wg.Done()

			o.lc.Debugf("Executing '%s' action on %s", operation.Action, operation.ServiceName)
			_, err := o.executor(o.executorPath, operation.ServiceName, operation.Action)
			if err != nil {
				responses = append(responses, common.NewBaseResponse(operation.RequestId, err.Error(), http.StatusInternalServerError))
				return
			}

			responses = append(responses, common.NewBaseResponse(operation.RequestId, "", http.StatusOK))
		}(operation)
	}

	wg.Wait()
	return responses, nil
}
