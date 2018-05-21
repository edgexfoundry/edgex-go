/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *******************************************************************************/
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/core/data"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/heartbeat"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
)

var loggingClient logger.LoggingClient

func main() {
	start := time.Now()
	var useConsul bool
	var useProfile string

	flag.BoolVar(&useConsul, "consul", false, "Indicates the service should use consul.")
	flag.BoolVar(&useConsul, "c", false, "Indicates the service should use consul.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	//Read Configuration
	configuration := &data.ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		logBeforeTermination(err)
		return
	}

	//Determine if configuration should be overridden from Consul
	var consulMsg string
	if useConsul {
		consulMsg = "Loading configuration from Consul..."
		err := data.ConnectToConsul(*configuration)
		if err != nil {
			logBeforeTermination(err)
			return //end program since user explicitly told us to use Consul.
		}
	} else {
		consulMsg = "Bypassing Consul configuration..."
	}

	// Setup Logging
	logTarget := setLoggingTarget(*configuration)
	loggingClient = logger.NewClient(internal.CoreDataServiceKey, configuration.EnableRemoteLogging, logTarget)

	loggingClient.Info(consulMsg)
	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.CoreDataServiceKey, edgex.Version))

	err = data.Init(*configuration, loggingClient, useConsul)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("call to init() failed: %v", err.Error()))
		return
	}

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.ServiceTimeout), "Request timed out")
	loggingClient.Info(configuration.AppOpenMsg, "")

	heartbeat.Start(configuration.HeartBeatMsg, configuration.HeartBeatTime, loggingClient)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHttpServer(errs, configuration.ServicePort)

	// Time it took to start service
	loggingClient.Info("Service started in: "+time.Since(start).String(), "")
	loggingClient.Info("Listening on port: " + strconv.Itoa(configuration.ServicePort))
	c := <-errs
	data.Destruct()
	loggingClient.Warn(fmt.Sprintf("terminating: %v", c))
}

func logBeforeTermination(err error) {
	loggingClient = logger.NewClient(internal.CoreDataServiceKey, false, "")
	loggingClient.Error(err.Error())
}

func setLoggingTarget(conf data.ConfigurationStruct) string {
	logTarget := conf.LoggingRemoteURL
	if !conf.EnableRemoteLogging {
		return conf.LoggingFile
	}
	return logTarget
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

func startHttpServer(errChan chan error, port int) {
	go func() {
		r := data.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), r)
	}()
}