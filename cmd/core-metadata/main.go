/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

var loggingClient logger.LoggingClient
var serviceTimeout int = 30000

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

	bootstrap(useConsul, useProfile)

	// Setup Logging
	logTarget := setLoggingTarget(*metadata.Configuration)
	loggingClient = logger.NewClient(internal.CoreMetaDataServiceKey, metadata.Configuration.EnableRemoteLogging, logTarget)

	loggingClient.Info("Service dependencies resolved...")
	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.CoreMetaDataServiceKey, edgex.Version))

	err := metadata.Init(loggingClient)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("call to init() failed: %v", err.Error()))
		return
	}

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(metadata.Configuration.ServiceTimeout), "Request timed out")
	loggingClient.Info(metadata.Configuration.AppOpenMsg, "")

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHttpServer(errs, metadata.Configuration.ServicePort)

	// Time it took to start service
	loggingClient.Info("Service started in: "+time.Since(start).String(), "")
	loggingClient.Info("Listening on port: " + strconv.Itoa(metadata.Configuration.ServicePort))
	c := <-errs
	metadata.Destruct()
	loggingClient.Warn(fmt.Sprintf("terminating: %v", c))
}

func logBeforeTermination(err error) {
	loggingClient = logger.NewClient(internal.CoreMetaDataServiceKey, false, "")
	loggingClient.Error(err.Error())
}

func setLoggingTarget(conf metadata.ConfigurationStruct) string {
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
		r := metadata.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), r)
	}()
}

func bootstrap(useConsul bool, useProfile string) {
	deps := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go metadata.ResolveDependencies(useConsul, useProfile, serviceTimeout, &wg, deps)
	go func(ch chan error) {
		for {
			select {
			case e, ok := <-ch:
				if ok {
					logBeforeTermination(e)
				} else {
					return
				}
			}
		}
	}(deps)

	wg.Wait()
}
