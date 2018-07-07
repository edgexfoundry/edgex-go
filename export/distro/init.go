//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/support/consul-client"
	"go.uber.org/zap"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/core/clients/coredata"
	"github.com/edgexfoundry/edgex-go/core/clients/types"
)

const (
	PingApiPath  = "/api/v1/ping"
	EventUriPath = "/api/v1/event"
)

var logger *zap.Logger

func ConnectToConsul(conf ConfigurationStruct) error {
	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    internal.ExportDistroServiceKey,
		ServicePort:    conf.ConsulPort,
		ServiceAddress: conf.ConsulHost,
		CheckAddress:   "http://" + conf.Hostname + ":" + strconv.Itoa(conf.Port) + PingApiPath,
		CheckInterval:  conf.CheckInterval,
		ConsulAddress:  conf.ConsulHost,
		ConsulPort:     conf.ConsulPort,
	})

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(&conf, internal.ExportDistroServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func Init(conf ConfigurationStruct, l *zap.Logger) error {
	configuration = conf
	logger = l

	coreDataEventURL := "http://" + conf.DataHost + ":" + strconv.Itoa(conf.DataPort) + EventUriPath

	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        EventUriPath,
		UseRegistry: false,
		Url:         coreDataEventURL,
	}

	ec = coredata.NewEventClient(params, nil)

	return nil
}
