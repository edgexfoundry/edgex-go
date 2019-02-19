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
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
)

func main() {

	start := time.Now()
	var useConsul bool
	var useProfile string

	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()
	useConsul = false

	params := startup.BootParams{UseConsul: useConsul, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, agent.Retry, logBeforeInit)

	ok := agent.Init()
	if !ok {
		logBeforeInit(fmt.Errorf("%s: service bootstrap failed", internal.SystemManagementAgentServiceKey))
		os.Exit(1)
	}

	logs.LoggingClient.Info("Service dependencies resolved...")
	logs.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.SystemManagementAgentServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(agent.Configuration.ServiceTimeout), "Request timed out")
	logs.LoggingClient.Info(agent.Configuration.AppOpenMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHttpServer(errs, agent.Configuration.ServicePort)

	// Time it took to start service
	logs.LoggingClient.Info("Service started in: " + time.Since(start).String())
	logs.LoggingClient.Info("Listening on port: " + strconv.Itoa(agent.Configuration.ServicePort))
	c := <-errs
	logs.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.SystemManagementAgentServiceKey, false, "", logger.InfoLog)
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
		correlation.LoggingClient = logs.LoggingClient
		r := agent.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), r)
	}()
}
