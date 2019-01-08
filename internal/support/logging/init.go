//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/pkg/errors"
)

var Configuration *ConfigurationStruct
var dbClient persistence
var LoggingClient logger.LoggingClient
var chConfig chan interface{} //A channel for use by ConsulDecoder in detecting configuration mods.

func Retry(useConsul bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	LoggingClient = newPrivateLogger()
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
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
			}
		}

		//Only attempt to connect to database if configuration has been populated
		if Configuration != nil {
			err = getPersistence()
			if err != nil {
				ch <- err
			} else {
				break
			}
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func Init(useConsul bool) bool {
	if Configuration == nil || dbClient == nil {
		return false
	}
	if useConsul {
		chConfig = make(chan interface{})
		go listenForConfigChanges()
	}
	return true
}

func Destruct() {
	if dbClient != nil {
		dbClient.closeSession()
		dbClient = nil
	}
	if chConfig != nil {
		close(chConfig)
		chConfig = nil
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
	cfg := consulclient.NewConsulConfig(conf.Registry, conf.Service, internal.SupportLoggingServiceKey)
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
	dec.Prefix = internal.ConfigRegistryStem + internal.SupportLoggingServiceKey
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
	dec.Prefix = internal.ConfigRegistryStem + internal.SupportLoggingServiceKey
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

func getPersistence() error {
	switch Configuration.Persistence {
	case PersistenceFile:
		dbClient = &fileLog{filename: Configuration.Logging.File}
	case PersistenceDB:
		// TODO: Integrate db layer with internal/pkg/db/ types so we can support other databases
		ms, err := connectToMongo()
		if err != nil {
			return err
		} else {
			dbClient = &mongoLog{session: ms}
		}
	default:
		return errors.New(fmt.Sprintf("unrecognized value Configuration.Persistence: %s", Configuration.Persistence))
	}
	return nil
}
