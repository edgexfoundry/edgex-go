//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/clients/scheduler"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

var loggingClient logger.LoggingClient
var ticker          = time.NewTicker(ScheduleInterval * time.Millisecond)
var schedulerClient scheduler.SchedulerClient

func ConnectToConsul(conf ConfigurationStruct) error {

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    internal.SupportSchedulerServiceKey,
		ServicePort:    conf.ServicePort,
		ServiceAddress: conf.ConsulHost,
		CheckAddress:   "http://" + conf.ConsulHost + ":" + strconv.Itoa(conf.ConsulPort) + PingApiPath,
		CheckInterval:  conf.CheckInterval,
		ConsulAddress:  conf.ConsulHost,
		ConsulPort:     conf.ConsulPort,
	})

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(&conf, internal.CoreCommandServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func Init(conf ConfigurationStruct, sc scheduler.SchedulerClient, l logger.LoggingClient, useConsul bool) error {
	loggingClient = l
	configuration = conf

	//TODO: The above two are set due to global scope throughout the package. How can this be eliminated / refactored?
	// Create scheduler client
    schedulerClient = sc
	ticker = time.NewTicker(time.Duration(conf.ScheduleInterval) * time.Millisecond)

	return nil
}

