//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/support/consul-client"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

const (
	PingApiPath = "/api/v1/ping"
)

var logger *zap.Logger

func ConnectToConsul(conf ConfigurationStruct) error {
	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    conf.ApplicationName,
		ServicePort:    conf.Port,
		ServiceAddress: conf.Hostname,
		CheckAddress:   "http://" + conf.Hostname + ":" + strconv.Itoa(conf.Port) + PingApiPath,
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

func Init(conf ConfigurationStruct, l *zap.Logger) error {
	configuration = conf
	logger = l

	return nil
}
