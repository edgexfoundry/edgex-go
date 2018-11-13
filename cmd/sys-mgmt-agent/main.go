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
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
)

func main() {

	// TODO: 1.  What kind of discoverability scenarios should be considered?
	// TODO: 2.  What kind of 'interaction' scenarios (involving Vault and Kong) should be considered?
	// TODO: 3.  Allow the agent/APIs to be turned off (or not used)...
	// TODO: 4.  Each service containing the SM API must have a configuration setting that turns off, protects or otherwise causes the SM API on the services to no-op.
	// TODO: 5.  When (if!) Consul is running, the SMA could request Consul to provide a catalog of registered services. To remove reliance on Consul, the SMA will be provided configuration (a manifest) that specifies the services it is to manage (Thus, the SMA can bootstrap all of EdgeX!)
	// TODO: 6.  Add a bootstrap property akin to –consul (e.g. –sma) to indicate directive for fetching the configuration from the local file system.
	// TODO: 7.  Effectively, the SMA can either (1) get configuration from Consul, or (2) get configuration from the local file system. This will be a manifest configuration file.
	// TODO: 8.  This file will boostrap the SMA. Different versions of the file may exist depending on how/where EdgeX is deployed (Docker v. Snappy, Windows v. Linux, etc.).
	// TODO: 9.  System Of Record (SOR) for configuration: The SMA must request configuration information from the _service_ itself!
	// TODO: 10. And BTW, what of the SMA's managing 3rd party infrastructure services (MongoDB, Consul, Kong, etc.) that will _not_ adhere to EdgeX guidelines?

	start := time.Now()
	var useConsul bool
	var useProfile string

	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()
	// [1] Removed two lines of code above (having to do with accepting the "--consul" aka "--c" flag) because the SMA is
	// designed to operate independently, in a stand-alone fashion. In particular, the SMA should not rely on Consul,
	// and leaving the --consul flag (in the relevant Docker file) was an oversight, as was originally the case (but
	// corrected since then by updating that Docker file).
	// [2] Added one line of code below (having to do with setting the "useConsul" variable to false) since we no longer
	// handle the --consul flag.
	useConsul = false

	params := startup.BootParams{UseConsul: useConsul, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, agent.Retry, logBeforeInit)

	ok := agent.Init()
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", internal.SystemManagementAgentServiceKey))
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
