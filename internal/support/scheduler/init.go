//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2018 Dell Inc
//
// SPDX-License-Identifier: Apache-2.0
package scheduler

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/pkg/errors"
)

var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient
var msc metadata.ScheduleClient
var msec metadata.ScheduleEventClient
var mac metadata.AddressableClient

var chConfig chan interface{} //A channel for use by ConsulDecoder in detecting configuration mods.
var ticker = time.NewTicker(time.Duration(ScheduleInterval) * time.Millisecond)

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
				LoggingClient = logger.NewClient(internal.SupportSchedulerServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Logging.Level)

				//Initialize service clients
				initializeClients(useConsul)
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

	if ticker != nil {
		StopTicker()
	}
}

func initializeConfiguration(useConsul bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain ConsulHost/Port
	conf := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, conf)
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
		//Check that information was successfully read from Consul
		if conf.Service.Port == 0 {
			return nil, errors.New("error reading from Consul")
		}
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

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}

func initializeClients(useConsul bool) {
	// Create metadata clients
	params := types.EndpointParams{
		ServiceKey:  internal.SupportSchedulerServiceKey,
		Path:        clients.ApiScheduleRoute,
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Metadata"].Url() + clients.ApiScheduleRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}
	// metadata Schedule client
	msc = metadata.NewScheduleClient(params, startup.Endpoint{})

	// metadata ScheduleEvent client
	params.Path = clients.ApiScheduleEventRoute
	params.Url = Configuration.Clients["Metadata"].Url() + clients.ApiScheduleEventRoute
	msec = metadata.NewScheduleEventClient(params, startup.Endpoint{})

	// metadata Addressable client
	params.Path = clients.ApiAddressableRoute
	params.Url = Configuration.Clients["Metadata"].Url() + clients.ApiAddressableRoute
	mac = metadata.NewAddressableClient(params, startup.Endpoint{})
}
