//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net/http"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/container"
)

func GetConfigs(ctx context.Context, services []string, dic *di.Container) (res interface{}, err errors.EdgeX) {
	configs := make(map[string]common.ConfigResponse)
	rc := bootstrapContainer.RegistryFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	for _, service := range services {
		var client interfaces.GeneralClient
		if rc != nil {
			ok, err := rc.IsServiceAvailable(service)
			if err != nil {
				errMsg := fmt.Sprintf("service %s not found in Registry", service)
				return res, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, err)
			}
			if !ok {
				errMsg := fmt.Sprintf("service %s unavailable", service)
				return res, errors.NewCommonEdgeX(errors.KindServiceUnavailable, errMsg, nil)
			}

			lc.Debugf("Registry responded with %s serviceName available", service)

			// Since service is unknown to SMA, ask the Registry for a ServiceEndpoint
			e, err := rc.GetServiceEndpoint(service)
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
			configuration := container.ConfigurationFrom(dic.Get)
			clientInfo, ok := configuration.Clients[service]
			if ok {
				client = clients.NewGeneralClient(clientInfo.Url())
			} else {
				errMsg := fmt.Sprintf("service %s not found in Configuration.Clients section", service)
				return res, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, err)
			}
		}

		config, err := client.FetchConfiguration(ctx)
		if err != nil {
			return res, errors.NewCommonEdgeXWrapper(err)
		}
		configs[service] = config
	}

	res = common.NewMultiConfigsResponse("", "", http.StatusOK, configs)
	return res, nil
}
