//
// Copyright (c) 2018
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/pkg/config"
	"github.com/edgexfoundry/edgex-go/pkg/usage"
	"github.com/edgexfoundry/edgex-go/support/logging"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
)

var loggingClient logger.LoggingClient

func main() {
	var useConsul, useProfile string

	flag.StringVar(&useConsul, "consul", "n", "Should the service use consul?")
	flag.StringVar(&useConsul, "c", "n", "Should the service use consul?")
	flag.StringVar(&useProfile, "profile", "default", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "default", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()
	flag.Parse()

	configuration := &logging.ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		logBeforeTermination(err)
		return
	}

	//Determine if configuration should be overridden from Consul
	var consulMsg string
	if useConsul == "y" {
		consulMsg = "Loading configuration from Consul..."
		err := logging.ConnectToConsul(*configuration)
		if err != nil {
			logBeforeTermination(err)
			return //end program since user explicitly told us to use Consul.
		}
	} else {
		consulMsg = "Bypassing Consul configuration..."
	}

	loggingClient = logger.NewClient(configuration.ApplicationName, false, configuration.LoggingFile)
	loggingClient.Info(consulMsg)

	logging.Init(*configuration)

	loggingClient.Info(fmt.Sprintf("Starting support-logging %s\n", edgex.Version))
	errs := make(chan error, 2)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logging.StartHTTPServer(errs)

	c := <-errs
	fmt.Println("terminated: ", c)
}

func logBeforeTermination(err error) {
	loggingClient = logger.NewClient(logging.SUPPORTLOGGINGSERVICENAME, false, "")
	loggingClient.Error(err.Error())
}
