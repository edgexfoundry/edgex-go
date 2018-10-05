//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2017 Dell Inc
//
// SPDX-License-Identifier: Apache-2.0
package scheduler

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/scheduler"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/pkg/errors"
)

var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient

var mdc metadata.DeviceClient
var cc metadata.CommandClient
var chConfig chan interface{} //A channel for use by ConsulDecoder in detecting configuration mods.


var ticker = time.NewTicker(time.Duration(ScheduleInterval) * time.Millisecond)
var schedulerClient scheduler.SchedulerClient

func Retry(useConsul bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	now := time.Now()
	until := now.Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(useConsul, useProfile)
			if err != nil {
				ch <- err
				if !useConsul {
					//Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			} else {
				//Check against boot timeout default
				if Configuration.Service.BootTimeout != timeout {
					until = now.Add(time.Millisecond * time.Duration(Configuration.Service.BootTimeout))
				}
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.SupportSchedulerServiceKey, Configuration.Logging.EnableRemote, logTarget)
			}
		}

		if Configuration != nil {
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func Init(useConsul bool) bool {
	if Configuration == nil {
		return false
	}

	if useConsul {
		chConfig = make(chan interface{})
		go listenForConfigChanges()
	}
	return true
}

func Destruct() {
	if chConfig != nil {
		close(chConfig)
	}
}

func initializeConfiguration(useConsul bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain ConsulHost/Port
	conf := &ConfigurationStruct{}
	err := config.LoadFromFileV2(useProfile, conf)
	if err != nil {
		return nil, err
	}

	if useConsul {
		conf, err = connectToConsul(conf)
		if err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func connectToConsul(conf *ConfigurationStruct) (*ConfigurationStruct, error) {
	//Obtain ConsulConfig
	cfg := consulclient.NewConsulConfig(conf.Registry, conf.Service, internal.SupportSchedulerServiceKey)

	// Register the service in Consul
	err := consulclient.ConsulInit(cfg)

	if err != nil {
		return conf, fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	}
	// Update configuration data from Consul
	updateCh := make(chan interface{})
	errCh := make(chan error)
	dec := consulclient.NewConsulDecoder(conf.Registry)
	dec.Target = &ConfigurationStruct{}
	dec.Prefix = internal.ConfigV2Stem + internal.SupportSchedulerServiceKey
	dec.ErrCh = errCh
	dec.UpdateCh = updateCh

	defer dec.Close()
	defer close(updateCh)
	defer close(errCh)
	go dec.Run()

	select {
	case <-time.After(2 * time.Second):
		err = errors.New("timeout loading config from registry")
	case ex := <-errCh:
		err = errors.New(ex.Error())
	case raw := <-updateCh:
		actual, ok := raw.(*ConfigurationStruct)
		if !ok {
			return conf, errors.New("type check failed")
		}
		conf = actual
	}

	return conf, err
}

func listenForConfigChanges() {
	errCh := make(chan error)
	dec := consulclient.NewConsulDecoder(Configuration.Registry)
	dec.Target = &ConfigurationStruct{}
	dec.Prefix = internal.ConfigV2Stem + internal.SupportSchedulerServiceKey
	dec.ErrCh = errCh
	dec.UpdateCh = chConfig

	defer dec.Close()
	defer close(errCh)

	go dec.Run()
	for {
		select {
		case ex := <-errCh:
			LoggingClient.Error(ex.Error())
		case raw, ok := <-chConfig:
			if ok {
				actual, ok := raw.(*ConfigurationStruct)
				if !ok {
					LoggingClient.Error("listenForConfigChanges() type check failed")
				}
				Configuration = actual //Mutex needed?
			} else {
				return
			}
		}
	}
}

func initializeClients(useConsul bool) {
	// Create metadata clients
	params := types.EndpointParams{
		ServiceKey:  internal.SupportSchedulerServiceKey,
		Path:        clients.ApiDeviceRoute,
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}

	mdc = metadata.NewDeviceClient(params, startup.Endpoint{})
	params.Path = clients.ApiCommandRoute
	params.Url = Configuration.Clients["Metadata"].Url() + clients.ApiCommandRoute
	cc = metadata.NewCommandClient(params, startup.Endpoint{})
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}

func AddDefaultSchedulers() error {

	LoggingClient.Info(fmt.Sprintf("loading default schedules and schedule events..."))

	defaultSchedules := Configuration.Schedules

	for i := range defaultSchedules {

		defaultSchedule := models.Schedule{
			BaseObject: models.BaseObject{},
			Id:         bson.NewObjectId(),
			Name:       defaultSchedules[i].Name,
			Start:      defaultSchedules[i].Start,
			End:        defaultSchedules[i].End,
			Frequency:  defaultSchedules[i].Frequency,
			Cron:       defaultSchedules[i].Cron,
			RunOnce:    defaultSchedules[i].RunOnce,
		}
		err := addSchedule(defaultSchedule)
		if err != nil {
			return LoggingClient.Error("AddDefaultSchedulers() - failed to load schedule %s", err.Error())
		} else {
			LoggingClient.Info(fmt.Sprintf("added default schedule %s", defaultSchedule.Name))
		}
	}

	defaultScheduleEvents := Configuration.ScheduleEvents

	for e := range defaultScheduleEvents {

		addressable := models.Addressable{
			// TODO: find a better way to initialize perhaps core-metadata
			Id:         bson.NewObjectId(),
			Name:       fmt.Sprintf("Schedule-%s", defaultScheduleEvents[e].Name),
			Path:       defaultScheduleEvents[e].Path,
			Port:       defaultScheduleEvents[e].Port,
			Protocol:   defaultScheduleEvents[e].Protocol,
			HTTPMethod: defaultScheduleEvents[e].Method,
			Address:    defaultScheduleEvents[e].Host,
		}

		defaultScheduleEvent := models.ScheduleEvent{
			Id: 	 	 bson.NewObjectId(),
			Name:        defaultScheduleEvents[e].Name,
			Schedule:    defaultScheduleEvents[e].Schedule,
			Parameters:  defaultScheduleEvents[e].Parameters,
			Service:     defaultScheduleEvents[e].Service,
			Addressable: addressable,
		}

		err := addScheduleEvent(defaultScheduleEvent)
		if err != nil {
			return LoggingClient.Error("AddDefaultSchedulers() - failed to load schedule event %s", err.Error())
		} else {
			LoggingClient.Info(fmt.Sprintf("added default schedule event %s", defaultScheduleEvent.Name))
		}
	}

	LoggingClient.Info(fmt.Sprintf("completed loading default schedules and schedule events"))
	return nil
}