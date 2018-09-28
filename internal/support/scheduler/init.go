//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/clients/scheduler"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var loggingClient logger.LoggingClient
var ticker = time.NewTicker(ScheduleInterval * time.Millisecond)
var schedulerClient scheduler.SchedulerClient
var initializeAttempts = 0

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

func Init(conf ConfigurationStruct, l logger.LoggingClient, useConsul bool) error {

	loggingClient = l
	configuration = conf

	// Check if we have default schedules to add
	if len(configuration.DefaultScheduleName) > 0  {
		// Add default scheduled events
		err := AddDefaultSchedules(configuration)
		if err != nil{
			return loggingClient.Error("Error while loading default schedule(s) or scheduleEvent(s) %s",err.Error())
		}
	}

	// TODO: Enable MetaData Client

	// Start ticker ('legacy')
	ticker = time.NewTicker(time.Duration(conf.ScheduleInterval) * time.Millisecond)

	return nil
}

func AddDefaultSchedules(conf ConfigurationStruct)  error{

	// Default number of attempts
	initializeAttempts++

	loggingClient.Info("bootstrapping default schedule attempt " + strconv.Itoa(initializeAttempts))

	// Add Schedule
	defaultSchedule := models.Schedule{
		BaseObject: models.BaseObject{},
		Id:         bson.NewObjectId(),
		Name:       conf.DefaultScheduleName,
		Start:      conf.DefaultScheduleStart,
		End:        "",
		Frequency:  conf.DefaultScheduleFrequency,
		Cron:       "",
		RunOnce:    false,
	}

	err := addSchedule(defaultSchedule)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("call to AddSchedule failed: %v", err.Error()))
	} else {
		loggingClient.Info("added default schedule " + conf.DefaultScheduleName)
	}

	// TODO: Change to using V2 Config where we can have map[string]

	// Add ScheduleEvent(s)
	eNames := strings.Split(conf.DefaultScheduleEventName, ",")
	eSchedules := strings.Split(conf.DefaultScheduleEventSchedule, ",")
	eParameters := strings.Split(conf.DefaultScheduleEventParameters, ",")
	eServices := strings.Split(conf.DefaultScheduleEventService, ",")
	ePaths := strings.Split(conf.DefaultScheduleEventPath, ",")
	eMethods := strings.Split(conf.DefaultScheduleEventMethod, ",")

	for i := range eNames {
		defScheduleEvent := models.ScheduleEvent{}

		defScheduleEvent.Name = eNames[i]
		defScheduleEvent.Schedule = eSchedules[i]
		defScheduleEvent.Parameters = eParameters[i]
		defScheduleEvent.Service = eServices[i]

		// TODO: the existing java scheduler utilizes a device client.
		// TODO: this device client is used to create the "addressable" based on the service name.
		addressable := models.Addressable{}
		addressable.Name = "Schedule-" + eNames[i]
		addressable.Path = ePaths[i]
		addressable.Port = conf.DefaultScheduleServicePort
		addressable.Protocol = conf.DefaultScheduleServiceProtocol
		addressable.HTTPMethod = eMethods[i]
		addressable.Address = conf.DefaultSchedulerServiceAddress

		// TODO: need a better method to create unique addressable.Id values  (java version uses client utility to obtain the objectId)
		addressable.Id = bson.NewObjectId() // will be unique every time system starts up not desirable

		defScheduleEvent.Id = bson.NewObjectId()
		defScheduleEvent.Addressable = addressable

		err := addScheduleEvent(defScheduleEvent)
		if err != nil {
			loggingClient.Error(fmt.Sprintf("call to AddSchuldedEvent failed: %v", err.Error()))
		} else {
			loggingClient.Info("added default schedule " + eNames[i])
		}
	}
	return nil
}
