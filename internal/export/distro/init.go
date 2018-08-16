//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

const (
	PingApiPath  = "/api/v1/ping"
	EventUriPath = "/api/v1/event"
)

var logger *zap.Logger
var ec coredata.EventClient

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

func Init(conf ConfigurationStruct, l *zap.Logger, useConsul bool) error {
	configuration = conf
	logger = l

	coreDataEventURL := "http://" + conf.DataHost + ":" + strconv.Itoa(conf.DataPort) + EventUriPath

	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        EventUriPath,
		UseRegistry: useConsul,
		Url:         coreDataEventURL,
	}

	ec = coredata.NewEventClient(params, types.Endpoint{})

	return nil
}
