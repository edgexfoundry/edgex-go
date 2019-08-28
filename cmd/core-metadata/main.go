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
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/cmd/common"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {
	start := time.Now()
	var useRegistry bool
	var useProfile string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseRegistry: useRegistry, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, metadata.Retry, logBeforeInit)

	ok := metadata.Init(useRegistry)
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", clients.CoreMetaDataServiceKey))
		os.Exit(1)
	}

	metadata.LoggingClient.Info("Service dependencies resolved...")
	metadata.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", clients.CoreMetaDataServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(metadata.Configuration.Service.Timeout), "Request timed out")
	metadata.LoggingClient.Info(metadata.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	url := metadata.Configuration.Service.Host + ":" + strconv.Itoa(metadata.Configuration.Service.Port)
	common.StartHTTPServer(metadata.LoggingClient, metadata.Configuration.Service.Timeout, metadata.LoadRestRoutes(), url, errs)

	// Time it took to start service
	metadata.LoggingClient.Info("Service started in: " + time.Since(start).String())
	metadata.LoggingClient.Info("Listening on port: " + strconv.Itoa(metadata.Configuration.Service.Port))
	c := <-errs
	metadata.Destruct()
	metadata.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient(clients.CoreMetaDataServiceKey, false, "", models.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
