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

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/memory"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"

	"go.uber.org/zap"
)

const (
	PingApiPath = "/api/v1/ping"
)

// Global variables
var dbc export.DBClient
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

	// Create a database client
	dbConfig := db.Configuration{
		Host:         conf.MongoURL,
		Port:         conf.MongoPort,
		Timeout:      conf.MongoConnectTimeout,
		DatabaseName: conf.MongoDatabaseName,
		Username:     conf.MongoUsername,
		Password:     conf.MongoPassword,
	}
	var err error
	dbc, err = newDBClient(conf.DBType, dbConfig)
	if err != nil {
		dbc = nil
		return fmt.Errorf("couldn't create database client: %v", err.Error())
	}

	// Connect to the database
	err = dbc.Connect()
	if err != nil {
		dbc = nil
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

// Return the dbClient interface
func newDBClient(dbType string, config db.Configuration) (export.DBClient, error) {
	switch dbType {
	case db.MongoDB:
		return mongo.NewClient(config), nil
	case db.MemoryDB:
		return &memory.MemDB{}, nil
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}
