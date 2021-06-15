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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

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
func (o *operation) Do(_ context.Context, operations []requests.OperationRequest) ([]common.BaseWithServiceNameResponse, errors.EdgeX) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var responses []common.BaseWithServiceNameResponse

	for _, operation := range operations {
		wg.Add(1)
		go func(operation requests.OperationRequest) {
			defer wg.Done()

			o.lc.Debugf("Executing '%s' action on %s", operation.Action, operation.ServiceName)
			res, err := o.executor(o.executorPath, edgexPrefix+operation.ServiceName, operation.Action)
			if err != nil {
				mu.Lock()
				responses = append(responses, common.BaseWithServiceNameResponse{
					BaseResponse: common.NewBaseResponse(operation.RequestId, res, http.StatusInternalServerError),
					ServiceName:  operation.ServiceName,
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			responses = append(responses, common.BaseWithServiceNameResponse{
				BaseResponse: common.NewBaseResponse(operation.RequestId, "", http.StatusOK),
				ServiceName:  operation.ServiceName,
			})
			mu.Unlock()
		}(operation)
	}

	wg.Wait()
	return responses, nil
}
