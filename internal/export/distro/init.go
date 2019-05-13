//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"
	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

var LoggingClient logger.LoggingClient
var ec coredata.EventClient
var Configuration *ConfigurationStruct
var registryClient registry.Client
var registryErrors chan error        //A channel for "config wait errors" sourced from Registry
var registryUpdates chan interface{} //A channel for "config updates" sourced from Registry
var messageClient messaging.MessageClient
var messageErrors chan error
var messageEnvelopes chan *msgTypes.MessageEnvelope
var processStop chan bool

func Retry(useRegistry bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(useRegistry, useProfile)
			if err != nil {
				ch <- err
				if !useRegistry {
					//Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.ExportDistroServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				//Initialize service clients
				initializeClients(useRegistry)
			}
		} else {
			// once config is initialized, stop looping
			break
		}

		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func Init(useRegistry bool) bool {
	if Configuration == nil {
		return false
	}
	if useRegistry {
		registryErrors = make(chan error)
		registryUpdates = make(chan interface{})
		go listenForConfigChanges()
	}

	var err error
	messageErrors, messageEnvelopes, err = initMessaging(messageClient)
	processStop = make(chan bool)

	if err != nil {
		LoggingClient.Error(err.Error())
		return false
	}

	go telemetry.StartCpuUsageAverage()

	return true
}

func Destruct() {
	if registryErrors != nil {
		close(registryErrors)
	}

	if registryUpdates != nil {
		close(registryUpdates)
	}

	if processStop != nil {
		processStop <- true
		close(processStop)
	}

	if messageErrors != nil {
		close(messageErrors)
	}

	if messageEnvelopes != nil {
		close(messageEnvelopes)
	}
}

func connectToRegistry(conf *ConfigurationStruct) error {
	var err error
	registryConfig := registryTypes.Config{
		Host:            conf.Registry.Host,
		Port:            conf.Registry.Port,
		Type:            conf.Registry.Type,
		ServiceKey:      internal.ExportDistroServiceKey,
		ServiceHost:     conf.Service.Host,
		ServicePort:     conf.Service.Port,
		ServiceProtocol: conf.Service.Protocol,
		CheckInterval:   conf.Service.CheckInterval,
		CheckRoute:      clients.ApiPingRoute,
		Stem:            internal.ConfigRegistryStem,
	}

	registryClient, err = registry.NewRegistryClient(registryConfig)
	if err != nil {
		return fmt.Errorf("connection to Registry could not be made: %v", err.Error())
	}

	// Check if registry service is running
	if !registryClient.IsAlive() {
		return fmt.Errorf("registry is not available")
	}

	// Register the service with Registry
	err = registryClient.Register()
	if err != nil {
		return fmt.Errorf("could not register service with Registry: %v", err.Error())
	}
	return nil
}

func initializeClients(useRegistry bool) {
	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: useRegistry,
		Url:         Configuration.Clients["CoreData"].Url() + clients.ApiEventRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}

	ec = coredata.NewEventClient(params, startup.Endpoint{RegistryClient: &registryClient})

	// Create the messaging client
	var err error
	messageClient, err = messaging.NewMessageClient(msgTypes.MessageBusConfig{
		SubscribeHost: msgTypes.HostInfo{
			Host:     Configuration.MessageQueue.Host,
			Port:     Configuration.MessageQueue.Port,
			Protocol: Configuration.MessageQueue.Protocol,
		},
		Type: Configuration.MessageQueue.Type,
	})

	if err != nil {
		LoggingClient.Error("failed to create messaging client: " + err.Error())
	}
}

func initializeConfiguration(useRegistry bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain RegistryHost/Port
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		return nil, err
	}

	if useRegistry {
		err = connectToRegistry(configuration)
		if err != nil {
			return nil, err
		}

		rawConfig, err := registryClient.GetConfiguration(configuration)
		if err != nil {
			return nil, fmt.Errorf("could not get configuration from Registry: %v", err.Error())
		}

		actual, ok := rawConfig.(*ConfigurationStruct)
		if !ok {
			return nil, fmt.Errorf("configuration from Registry failed type check")
		}

		configuration = actual

		// Check that information was successfully read from Registry
		if configuration.Service.Port == 0 {
			return nil, errors.New("error reading configuration from Registry")
		}
	}
	return configuration, nil
}

func listenForConfigChanges() {
	if registryClient == nil {
		LoggingClient.Error("listenForConfigChanges() registry client not set")
		return
	}

	registryClient.WatchForChanges(registryUpdates, registryErrors, &WritableInfo{}, internal.WritableKey)

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-signals:
			// Quietly and gracefully stop when SIGINT/SIGTERM received
			return
		case raw, ok := <-registryUpdates:
			if ok {
				actual, ok := raw.(*WritableInfo)
				if !ok {
					LoggingClient.Error("listenForConfigChanges() type check failed")
				}

				Configuration.Writable = *actual

				LoggingClient.Info("Writeable configuration has been updated from the Registry")
				LoggingClient.SetLogLevel(Configuration.Writable.LogLevel)
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
