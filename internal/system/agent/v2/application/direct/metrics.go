//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package direct

import (
	"context"
	"fmt"
	"net/http"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"

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

func (m *metrics) Get(ctx context.Context, services []string) (res interface{}, err errors.EdgeX) {
	metrics := make(map[string]interface{})

	for _, service := range services {
		var client interfaces.GeneralClient
		if m.rc != nil {
			ok, err := m.rc.IsServiceAvailable(service)
			if err != nil {
				errMsg := fmt.Sprintf("service %s not found in Registry", service)
				return res, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, err)
			}
			if !ok {
				errMsg := fmt.Sprintf("service %s unavailable", service)
				return res, errors.NewCommonEdgeX(errors.KindServiceUnavailable, errMsg, nil)
			}

			m.lc.Infof("Registry responded with %s serviceName available", service)

			// Since service is unknown to SMA, ask the Registry for a ServiceEndpoint
			e, err := m.rc.GetServiceEndpoint(service)
			if err != nil {
				return res, errors.NewCommonEdgeXWrapper(err)
			}

			clientInfo := bootstrapConfig.ClientInfo{
				Protocol: "http",
				Host:     e.Host,
				Port:     e.Port,
			}
			client = clients.NewGeneralClient(clientInfo.Url())
		} else {
			c, ok := m.config.Clients[service]
			if ok {
				client = clients.NewGeneralClient(c.Url())
			} else {
				errMsg := fmt.Sprintf("service %s not found in Configuration.Clients section", service)
				return res, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, err)
			}
		}

		m, err := client.FetchMetrics(ctx)
		if err != nil {
			return res, errors.NewCommonEdgeXWrapper(err)
		}
		metrics[service] = m
	}

	res = common.NewMultiMetricsResponse("", "", http.StatusOK, metrics)
	return res, nil
}
