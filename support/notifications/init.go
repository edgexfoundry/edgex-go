/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package notifications

import (
	"fmt"
	"strings"

	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/edgexfoundry/edgex-go/support/notifications/clients"
	"github.com/robfig/cron"
)

// Global variables
var dbc clients.DBClient
var loggingClient logger.LoggingClient
var normalDuration string
var resendDuration string
var criticalResendDelay int
var limitMax int
var resendLimit int
var smtpPort string
var smtpHost string
var smtpSender string
var smtpPassword string
var smtpSubject string

func ConnectToConsul(conf ConfigurationStruct) error {

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    conf.ServiceName,
		ServicePort:    conf.ServicePort,
		ServiceAddress: conf.ServiceAddress,
		CheckAddress:   conf.ConsulCheckAddress,
		CheckInterval:  conf.CheckInterval,
		ConsulAddress:  conf.ConsulHost,
		ConsulPort:     conf.ConsulPort,
	})

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	}
	if err := consulclient.CheckKeyValuePairs(&conf, conf.ServiceName, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
		return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
	}
	return nil
}

func Init(conf ConfigurationStruct, l logger.LoggingClient) error {
	loggingClient = l
	configuration = conf
	normalDuration = conf.SchedulerNormalDuration
	resendDuration = conf.SchedulerNormalResendDuration
	criticalResendDelay = conf.SchedulerCriticalResendDelay
	limitMax = conf.ReadMaxLimit
	resendLimit = conf.ResendLimit
	smtpPort = conf.SMTPPort
	smtpHost = conf.SMTPHost
	smtpSender = conf.SMTPSender
	smtpPassword = conf.SMTPPassword
	smtpSubject = conf.SMTPSubject

	var err error

	// Create a database client
	dbc, err = clients.NewDBClient(clients.DBConfiguration{
		DbType:            clients.MONGO,
		Host:              conf.MongoDBHost,
		Port:              conf.MongoDBPort,
		Timeout:           conf.MongoDBConnectTimeout,
		DatabaseName:      conf.MongoDatabaseName,
		Username:          conf.MongoDBUserName,
		Password:          conf.MongoDBPassword,
		ReadMax:           conf.ReadMaxLimit,
		ResendLimit:       conf.ResendLimit,
		CleanupDefaultAge: conf.CleanupDefaultAge,
	})
	if err != nil {
		return fmt.Errorf("couldn't connect to database: %v", err.Error())
	}

	initNormalDistribution()
	initNormalResend()
	return nil
}

func initNormalDistribution() {
	loggingClient.Debug("Normal distribution occuring on cron schedule: " + normalDuration)
	c := cron.New()
	c.AddFunc(normalDuration, func() { startNormalDistributing() })
	c.Start()
}

func initNormalResend() {
	loggingClient.Debug("Normal resend occuring on cron schedule: " + resendDuration)
	c := cron.New()
	c.AddFunc(resendDuration, func() { startNormalResend() })
	c.Start()
}
