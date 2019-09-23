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
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/support/logging"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {
	start := time.Now()
	var useRegistry bool
	var configDir, profileDir string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use Registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use Registry.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{
		UseRegistry: useRegistry,
		ConfigDir:   configDir,
		ProfileDir:  profileDir,
		BootTimeout: internal.BootTimeoutDefault,
	}
	startup.Bootstrap(params, logging.Retry, logBeforeInit)

	ok := logging.Init(useRegistry)
	if !ok {
		time.Sleep(time.Millisecond * time.Duration(15))
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", clients.SupportLoggingServiceKey))
		os.Exit(1)
	}
	logging.LoggingClient.Info("Service dependencies resolved...")
	logging.LoggingClient.Info(fmt.Sprintf("Starting %s %s", clients.SupportLoggingServiceKey, edgex.Version))

	errs := make(chan error, 2)
	listenForInterrupt(errs)

	// Time it took to start service
	logging.LoggingClient.Info("Service started in: " + time.Since(start).String())
	logging.LoggingClient.Info("Listening on port: " + strconv.Itoa(logging.Configuration.Service.Port))
	url := logging.Configuration.Service.Host + ":" + strconv.Itoa(logging.Configuration.Service.Port)
	startup.StartHTTPServer(logging.LoggingClient, logging.Configuration.Service.Timeout, logging.LoadRestRoutes(), url, errs)

	c := <-errs
	logging.Destruct()
	logging.LoggingClient.Warn(fmt.Sprintf("terminated %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient(clients.SupportLoggingServiceKey, false, "", models.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
