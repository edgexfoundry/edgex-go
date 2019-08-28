/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
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
 *
 * @microservice: support-notifications
 * @author: Jim White, Dell Technologies
 * @version: 0.5.0
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
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {
	start := time.Now()
	var useRegistry bool
	var useProfile string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use Registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use Registry.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseRegistry: useRegistry, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, notifications.Retry, logBeforeInit)

	ok := notifications.Init(useRegistry)
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", clients.SupportNotificationsServiceKey))
		os.Exit(1)
	}

	notifications.LoggingClient.Info("Service dependencies resolved...")
	notifications.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", clients.SupportNotificationsServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(notifications.Configuration.Service.Timeout), "Request timed out")
	notifications.LoggingClient.Info(notifications.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHTTPServer(errs)

	// Time it took to start service
	notifications.LoggingClient.Info("Service started in: " + time.Since(start).String())
	notifications.LoggingClient.Info("Listening on port: " + strconv.Itoa(notifications.Configuration.Service.Port))
	c := <-errs
	notifications.Destruct()
	notifications.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient(clients.SupportNotificationsServiceKey, false, "", models.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

func startHTTPServer(errChan chan error) {
	go func() {
		correlation.LoggingClient = notifications.LoggingClient //Not thrilled about this, can't think of anything better ATM
		timeout := time.Millisecond * time.Duration(notifications.Configuration.Service.Timeout)

		server := &http.Server{
			Handler:      notifications.LoadRestRoutes(),
			Addr:         notifications.Configuration.Service.Host + ":" + strconv.Itoa(notifications.Configuration.Service.Port),
			WriteTimeout: timeout,
			ReadTimeout:  timeout),
		}

		errChan <- server.ListenAndServe()
	}()
}
