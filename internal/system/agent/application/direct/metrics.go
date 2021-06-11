//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package direct

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"

	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/config"
)

type metrics struct {
	lc     logger.LoggingClient
	rc     registry.Client
	config *config.ConfigurationStruct
}

func NewMetrics(lc logger.LoggingClient, rc registry.Client, config *config.ConfigurationStruct) *metrics {
	return &metrics{
		lc:     lc,
		rc:     rc,
		config: config,
	}
}

func (m *metrics) Get(ctx context.Context, services []string) ([]interface{}, errors.EdgeX) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var responses []interface{}

	for _, service := range services {
		wg.Add(1)
		go func(serviceName string) {
			defer wg.Done()

			var client interfaces.GeneralClient
			if m.rc != nil {
				ok, err := m.rc.IsServiceAvailable(serviceName)
				if err != nil {
					mu.Lock()
					responses = append(responses, common.BaseWithMetricsResponse{
						BaseResponse: common.NewBaseResponse("", err.Error(), http.StatusNotFound),
						ServiceName:  serviceName,
						Metrics:      nil,
					})
					mu.Unlock()
					return
				}
				if !ok {
					errMsg := fmt.Sprintf("service %s unavailable", serviceName)
					mu.Lock()
					responses = append(responses, common.BaseWithMetricsResponse{
						BaseResponse: common.NewBaseResponse("", errMsg, http.StatusServiceUnavailable),
						ServiceName:  serviceName,
						Metrics:      nil,
					})
					mu.Unlock()
					return
				}

				m.lc.Debugf("Registry responded with %s service available", serviceName)

				// Since service is unknown to SMA, ask the Registry for a ServiceEndpoint
				e, err := m.rc.GetServiceEndpoint(serviceName)
				if err != nil {
					mu.Lock()
					responses = append(responses, common.BaseWithMetricsResponse{
						BaseResponse: common.NewBaseResponse("", err.Error(), http.StatusInternalServerError),
						ServiceName:  serviceName,
						Metrics:      nil,
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
				c, ok := m.config.Clients[serviceName]
				if ok {
					client = clients.NewGeneralClient(c.Url())
				} else {
					errMsg := fmt.Sprintf("service %s not found in Configuration.Clients section", serviceName)
					mu.Lock()
					responses = append(responses, common.BaseWithMetricsResponse{
						BaseResponse: common.NewBaseResponse("", errMsg, http.StatusNotFound),
						ServiceName:  serviceName,
						Metrics:      nil,
					})
					mu.Unlock()
					return
				}
			}

			m, err := client.FetchMetrics(ctx)
			if err != nil {
				mu.Lock()
				responses = append(responses, common.BaseWithMetricsResponse{
					BaseResponse: common.NewBaseResponse("", err.Error(), http.StatusInternalServerError),
					ServiceName:  serviceName,
					Metrics:      nil,
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			responses = append(responses, common.BaseWithMetricsResponse{
				BaseResponse: common.NewBaseResponse("", "", http.StatusOK),
				ServiceName:  serviceName,
				Metrics:      m,
			})
			mu.Unlock()
		}(service)
	}

	wg.Wait()
	return responses, nil
}
