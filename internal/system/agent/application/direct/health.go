//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package direct

import (
	"net/http"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"
)

func GetHealth(services []string, rc registry.Client) ([]common.BaseWithServiceNameResponse, errors.EdgeX) {
	if rc == nil {
		err := errors.NewCommonEdgeX(errors.KindServerError, "registry is required", nil)
		return nil, err
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	var responses []common.BaseWithServiceNameResponse

	for _, service := range services {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()

			// the registry service returns nil for a healthy service
			ok, err := rc.IsServiceAvailable(serviceName)
			if err != nil {
				mu.Lock()
				responses = append(responses, common.BaseWithServiceNameResponse{
					BaseResponse: common.NewBaseResponse("", err.Error(), http.StatusNotFound),
					ServiceName:  serviceName,
				})
				mu.Unlock()
				return
			}

			if !ok {
				mu.Lock()
				responses = append(responses, common.BaseWithServiceNameResponse{
					BaseResponse: common.NewBaseResponse("", "service unhealthy", http.StatusServiceUnavailable),
					ServiceName:  serviceName,
				})
				mu.Unlock()
				return
			}
			mu.Lock()
			responses = append(responses, common.BaseWithServiceNameResponse{
				BaseResponse: common.NewBaseResponse("", "", http.StatusOK),
				ServiceName:  serviceName,
			})
			mu.Unlock()
		}(service)
	}

	wg.Wait()
	return responses, nil
}
