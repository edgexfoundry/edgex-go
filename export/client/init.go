//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/export/client/clients"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"go.uber.org/zap"
)

const (
	PingApiPath = "/api/v1/ping"
)

// Global variables
var dbc clients.DBClient
var logger *zap.Logger

func ConnectToConsul(conf ConfigurationStruct) error {
	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    internal.ExportClientServiceKey,
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
		if err := consulclient.CheckKeyValuePairs(&conf, internal.ExportClientServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func Init(conf ConfigurationStruct, l *zap.Logger) error {
	configuration = conf
	logger = l

	var err error

	// Create a database client
	dbc, err = clients.NewDBClient(clients.DBConfiguration{
		DbType:       clients.GetDatabaseType(conf.DBType),
		Host:         conf.MongoURL,
		Port:         conf.MongoPort,
		Timeout:      conf.MongoConnectTimeout,
		DatabaseName: conf.MongoDatabaseName,
		Username:     conf.MongoUsername,
		Password:     conf.MongoPassword,
	})
	if err != nil {
		return fmt.Errorf("couldn't connect to database: %v", err.Error())
	}

	return nil
}

func Destroy() {
	if dbc != nil {
		dbc.CloseSession()
		dbc = nil
	}
}
