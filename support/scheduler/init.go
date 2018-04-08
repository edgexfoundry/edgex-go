//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/support/consul-client"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/edgexfoundry/edgex-go/support/scheduler-client"
	"strconv"
	"strings"
	"time"
)

const (
	PingApiPath = "/api/v1/ping"
)

var (
	loggingClient   logger.LoggingClient
	schedulerClient scheduler.SchedulerClient
	ticker          = time.NewTicker(ScheduleInterval * time.Millisecond)
)

func ConnectToConsul(conf ConfigurationStruct) error {
	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    conf.ServiceName,
		ServicePort:    conf.ServerPort,
		ServiceAddress: conf.ServiceHost,
		CheckAddress:   "http://" + conf.ServiceHost + ":" + strconv.Itoa(conf.ServerPort) + PingApiPath,
		CheckInterval:  conf.CheckInterval,
		ConsulAddress:  conf.ConsulHost,
		ConsulPort:     conf.ConsulPort,
	})

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(&conf, conf.ApplicationName, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func Init(conf ConfigurationStruct, l logger.LoggingClient, sc scheduler.SchedulerClient) {
	loggingClient = l
	schedulerClient = sc
	configuration = conf
	ticker = time.NewTicker(time.Duration(conf.ScheduleInterval) * time.Millisecond)
}
