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
	"github.com/edgexfoundry/edgex-go/internal/support/logging"
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
	startup.Bootstrap(params, logging.Retry, logBeforeInit)

	ok := logging.Init(useConsul)
	if !ok {
		time.Sleep(time.Millisecond * time.Duration(15))
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", internal.SupportLoggingServiceKey))
		return
	}
	logging.LoggingClient.Info("Service dependencies resolved...")
	logging.LoggingClient.Info(fmt.Sprintf("Starting %s %s", internal.SupportLoggingServiceKey, edgex.Version))

	errs := make(chan error, 2)
	listenForInterrupt(errs)

	// Time it took to start service
	logging.LoggingClient.Info("Service started in: " + time.Since(start).String())
	logging.LoggingClient.Info("Listening on port: " + strconv.Itoa(logging.Configuration.Service.Port))
	startHTTPServer(errs)

	c := <-errs
	logging.Destruct()
	logging.LoggingClient.Warn(fmt.Sprintf("terminated %v", c))
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.SupportLoggingServiceKey, false, "", logger.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

func startHTTPServer(errChan chan error) {
	go func() {
		p := fmt.Sprintf(":%d", logging.Configuration.Service.Port)
		errChan <- http.ListenAndServe(p, logging.HttpServer())
	}()
}
