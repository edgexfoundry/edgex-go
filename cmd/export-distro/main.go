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
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export/distro"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
)

func main() {
	var useRegistry bool
	var useProfile string
	start := time.Now()

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use Registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use Registry.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseRegistry: useRegistry, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, distro.Retry, logBeforeInit)

	if ok := distro.Init(useRegistry); !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed", internal.ExportDistroServiceKey))
		os.Exit(1)
	}

	distro.LoggingClient.Info("Service dependencies resolved...")
	distro.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.ExportDistroServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(distro.Configuration.Service.Timeout), "Request timed out")
	distro.LoggingClient.Info(distro.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)

	listenForInterrupt(errs)

	distro.Loop()

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
	l := logger.NewClient(internal.ExportDistroServiceKey, false, "", logger.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		correlation.LoggingClient = distro.LoggingClient
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
