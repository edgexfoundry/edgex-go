//
// Copyright (c) 2017
// Mainflux
// Cavium
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

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export/client"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
)

func main() {
	start := time.Now()
	var (
		useConsul  bool
		useProfile string
	)

	flag.BoolVar(&useConsul, "consul", false, "Indicates the service should use consul.")
	flag.BoolVar(&useConsul, "c", false, "Indicates the service should use consul.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseConsul: useConsul, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, client.Retry, logBeforeInit)

	ok := client.Init(useConsul)
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed", internal.ExportClientServiceKey))
		os.Exit(1)
	}

	client.LoggingClient.Info("Service dependencies resolved...")
	client.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.ExportClientServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(client.Configuration.Service.Timeout), "Request timed out")
	client.LoggingClient.Info(client.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	client.StartHTTPServer(errs)

	// Time it took to start service
	client.LoggingClient.Info("Service started in: " + time.Since(start).String())
	client.LoggingClient.Info("Listening on port: " + strconv.Itoa(client.Configuration.Service.Port))
	c := <-errs
	client.Destruct()
	client.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.ExportClientServiceKey, false, "", logger.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		correlation.LoggingClient = client.LoggingClient
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
