//
// Copyright (c) 2018

// Dell, Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"flag"
	"fmt"
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/pkg/config"
	"github.com/edgexfoundry/edgex-go/pkg/heartbeat"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/edgexfoundry/edgex-go/support/scheduler"
	sc "github.com/edgexfoundry/edgex-go/support/scheduler-client"
	"net/http"
	"strconv"
	"time"
)

var loggingClient logger.LoggingClient

func main() {
	start := time.Now()
	var (
		useConsul = flag.String("consul", "", "Should the service use consul?")
		useProfile = flag.String("profile", "default", "Specify a profile other than default.")
	)
	flag.Parse()

	//Read Configuration
	configuration := &scheduler.ConfigurationStruct{}
	err := config.LoadFromFile(*useProfile, configuration)
	if err != nil {
		logBeforeTermination(err)
		return
	}

	//Determine if configuration should be overridden from Consul
	var consulMsg string
	if *useConsul == "y" {
		consulMsg = "loading configuration from Consul..."
		err := scheduler.ConnectToConsul(*configuration)
		if err != nil {
			logBeforeTermination(err)
			return //end program since user explicitly told us to use Consul.
		}
	} else {
		consulMsg = "bypassing Consul configuration..."
	}

	// Setup Logging
	logTarget := setLoggingTarget(*configuration)
	var loggingClient = logger.NewClient(configuration.ApplicationName, configuration.EnableRemoteLogging, logTarget)

	loggingClient.Info(consulMsg)
	loggingClient.Info(fmt.Sprintf("starting %s %s ", scheduler.SupportScheduleServiceName, edgex.Version))

	var schedulerClient = sc.SchedulerClient{
		SchedulerServiceHost: configuration.ServiceHost,
		SchedulerServicePort: 48081,
		OwningService: scheduler.SupportScheduleServiceName,
	}

	scheduler.Init(*configuration, loggingClient, schedulerClient)

	r := scheduler.LoadRestRoutes()
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.ServerTimeout), "Request timed out")
	loggingClient.Info(configuration.AppOpenMsg, "")

	heartbeat.Start(configuration.HeartbeatMsg, configuration.HeartbeatTime, loggingClient)

	scheduler.StartTicker()

	// Time it took to start service
	loggingClient.Info("service started in: "+time.Since(start).String(), "")
	loggingClient.Info("listening on port: "+strconv.Itoa(configuration.ServerPort), "")
	loggingClient.Error(http.ListenAndServe(":"+strconv.Itoa(configuration.ServerPort), r).Error())

	scheduler.StopTicker()
}
func logBeforeTermination(err error) {
	loggingClient = logger.NewClient(scheduler.SupportScheduleServiceName, false, "")
	loggingClient.Error(err.Error())
}

func setLoggingTarget(conf scheduler.ConfigurationStruct) string {
	logTarget := conf.LoggingRemoteUrl
	if !conf.EnableRemoteLogging {
		return conf.LoggingFile
	}
	return logTarget
}