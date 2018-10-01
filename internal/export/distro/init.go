//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

const (
	PingApiPath = "/api/v1/ping"
)

var LoggingClient logger.LoggingClient
var ec coredata.EventClient
var Configuration *ConfigurationStruct

type ConfigurationStruct struct {
	Hostname             string
	Port                 int
	Timeout              int
	DistroHost           string
	ClientHost           string
	DataHost             string
	DataPort             int
	ConsulHost           string
	ConsulPort           int
	ConsulProfilesActive string
	CheckInterval        string
	MQTTSCert            string
	MQTTSKey             string
	MarkPushed           bool
	AWSCert              string
	AWSKey               string
	EnableRemoteLogging  bool
	LoggingRemoteURL     string
	LogFile              string
	AppOpenMsg           string
}

func Retry(useConsul bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
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
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.ExportDistroServiceKey, Configuration.EnableRemoteLogging, logTarget)
				//Initialize service clients
				initializeClient(useConsul)
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

func Init() bool {
	if Configuration == nil {
		return false
	}
	return true
}

func connectToConsul(conf *ConfigurationStruct) error {
	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    internal.ExportDistroServiceKey,
		ServicePort:    conf.ConsulPort,
		ServiceAddress: conf.ConsulHost,
		CheckAddress:   "http://" + conf.Hostname + ":" + strconv.Itoa(conf.Port) + PingApiPath,
		CheckInterval:  conf.CheckInterval,
		ConsulAddress:  conf.ConsulHost,
		ConsulPort:     conf.ConsulPort,
	})

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(conf, internal.ExportDistroServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func initializeClient(useConsul bool) {
	coreDataEventUrl := fmt.Sprintf("http://%s:%d%s", Configuration.DataHost, Configuration.DataPort, clients.ApiEventRoute)

	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: useConsul,
		Url:         coreDataEventUrl,
		Interval:    internal.ClientMonitorDefault,
	}

	ec = coredata.NewEventClient(params, startup.Endpoint{})
}

func initializeConfiguration(useConsul bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain ConsulHost/Port
	conf := &ConfigurationStruct{}
	if err := config.LoadFromFile(useProfile, conf); err != nil {
		return nil, err
	}

	if useConsul {
		if err := connectToConsul(conf); err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func setLoggingTarget() string {
	logTarget := Configuration.LoggingRemoteURL
	if !Configuration.EnableRemoteLogging {
		return Configuration.LogFile
	}
	return logTarget
}
