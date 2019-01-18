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
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

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

	params := startup.BootParams{UseConsul: useConsul, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, notifications.Retry, logBeforeInit)

	ok := notifications.Init(useConsul)
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", internal.SupportNotificationsServiceKey))
		os.Exit(1)
	}

	notifications.LoggingClient.Info("Service dependencies resolved...")
	notifications.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.SupportNotificationsServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(notifications.Configuration.Service.Timeout), "Request timed out")
	notifications.LoggingClient.Info(notifications.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHttpServer(errs, notifications.Configuration.Service.Port)

	// Time it took to start service
	notifications.LoggingClient.Info("Service started in: " + time.Since(start).String())
	notifications.LoggingClient.Info("Listening on port: " + strconv.Itoa(notifications.Configuration.Service.Port))
	c := <-errs
	notifications.Destruct()
	notifications.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.SupportNotificationsServiceKey, false, "", logger.InfoLog)
	l.Error(err.Error())
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
		r := notifications.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), r)
	}()
}
