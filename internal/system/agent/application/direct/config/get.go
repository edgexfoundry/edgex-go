//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
)

func GetConfigs(ctx context.Context, services []string, dic *di.Container) []common.BaseWithConfigResponse {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var responses []common.BaseWithConfigResponse
	rc := bootstrapContainer.RegistryFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	for _, service := range services {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()

			var client interfaces.GeneralClient
			if rc != nil {
				ok, err := rc.IsServiceAvailable(serviceName)
				if err != nil {
					mu.Lock()
					responses = append(responses, common.BaseWithConfigResponse{
						BaseResponse: common.NewBaseResponse("", err.Error(), http.StatusNotFound),
						ServiceName:  serviceName,
						Config:       nil,
					})
					mu.Unlock()
					return
				}
				if !ok {
					errMsg := fmt.Sprintf("service %s unavailable", serviceName)
					mu.Lock()
					responses = append(responses, common.BaseWithConfigResponse{
						BaseResponse: common.NewBaseResponse("", errMsg, http.StatusServiceUnavailable),
						ServiceName:  serviceName,
						Config:       nil,
					})
					mu.Unlock()
					return
				}

				lc.Debugf("Registry responded with %s service available", serviceName)

				// Since service is unknown to SMA, ask the Registry for a ServiceEndpoint
				e, err := rc.GetServiceEndpoint(serviceName)
				if err != nil {
					mu.Lock()
					responses = append(responses, common.BaseWithConfigResponse{
						BaseResponse: common.NewBaseResponse("", err.Error(), http.StatusInternalServerError),
						ServiceName:  serviceName,
						Config:       nil,
					})
					mu.Unlock()
					return
				}

				clientInfo := bootstrapConfig.ClientInfo{
					Protocol: "http",
					Host:     e.Host,
					Port:     e.Port,
				}
				client = clients.NewGeneralClient(clientInfo.Url())
			} else {
				configuration := container.ConfigurationFrom(dic.Get)
				clientInfo, ok := configuration.Clients[serviceName]
				if ok {
					client = clients.NewGeneralClient(clientInfo.Url())
				} else {
					errMsg := fmt.Sprintf("service %s not found in Configuration.Clients section", serviceName)
					mu.Lock()
					responses = append(responses, common.BaseWithConfigResponse{
						BaseResponse: common.NewBaseResponse("", errMsg, http.StatusNotFound),
						ServiceName:  serviceName,
						Config:       nil,
					})
					mu.Unlock()
					return
				}
			}

			config, err := client.FetchConfiguration(ctx)
			if err != nil {
				mu.Lock()
				responses = append(responses, common.BaseWithConfigResponse{
					BaseResponse: common.NewBaseResponse("", err.Error(), http.StatusInternalServerError),
					ServiceName:  serviceName,
					Config:       nil,
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			responses = append(responses, common.BaseWithConfigResponse{
				BaseResponse: common.NewBaseResponse("", "", http.StatusOK),
				ServiceName:  serviceName,
				Config:       config,
			})
			mu.Unlock()
		}(service)
	}

	wg.Wait()
	return responses
}
