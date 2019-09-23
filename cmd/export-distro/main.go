//
// Copyright (c) 2017
// Cavium
// Mainflux
// Copyright (c) 2019 Intel Corporation
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

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export/distro"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
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
	startup.Bootstrap(params, distro.Retry, logBeforeInit)

	if ok := distro.Init(useRegistry); !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed", clients.ExportDistroServiceKey))
		os.Exit(1)
	}

	distro.LoggingClient.Info("Service dependencies resolved...")
	distro.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", clients.ExportDistroServiceKey, edgex.Version))

	distro.LoggingClient.Info(distro.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)

	listenForInterrupt(errs)
	url := distro.Configuration.Service.Host + ":" + strconv.Itoa(distro.Configuration.Service.Port)
	startup.StartHTTPServer(distro.LoggingClient, distro.Configuration.Service.Timeout, distro.LoadRestRoutes(), url, errs)

	go distro.Loop()

	// Time it took to start service
	distro.LoggingClient.Info("Service started in: " + time.Since(start).String())
	distro.LoggingClient.Info("Listening on port: " + strconv.Itoa(distro.Configuration.Service.Port))
	c := <-errs
	distro.Destruct()
	distro.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	close(errs)
	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient(clients.ExportDistroServiceKey, false, "", models.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		correlation.LoggingClient = distro.LoggingClient
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
