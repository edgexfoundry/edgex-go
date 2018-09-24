//
// Copyright (c) 2017
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"flag"
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export/distro"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

var bootTimeout = 30000 //Once we start the V2 configuration rework, this will be config driven

func main() {
	var useConsul bool
	var useProfile string
	start := time.Now()

	flag.BoolVar(&useConsul, "consul", false, "Indicates the service should use consul.")
	flag.BoolVar(&useConsul, "c", false, "Indicates the service should use consul.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseConsul: useConsul, UseProfile: useProfile, BootTimeout: bootTimeout}
	startup.Bootstrap(params, distro.Retry, logBeforeInit)

	if ok := distro.Init(); !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed", internal.ExportDistroServiceKey))
		return
	}

	distro.LoggingClient.Info("Service dependencies resolved...")
	distro.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.ExportDistroServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(distro.Configuration.Timeout), "Request timed out")
	distro.LoggingClient.Info(distro.Configuration.AppOpenMsg, "")

	errs := make(chan error, 2)
	eventCh := make(chan *models.Event, 10)

	listenForInterrupt(errs)

	// There can be another receivers that can be initialized here
	distro.ZeroMQReceiver(eventCh)
	distro.Loop(errs, eventCh)

	// Time it took to start service
	distro.LoggingClient.Info("Service started in: "+time.Since(start).String(), "")
	distro.LoggingClient.Info("Listening on port: " + strconv.Itoa(distro.Configuration.Port))
	c := <-errs
	distro.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.ExportDistroServiceKey, false, "")
	l.Error(err.Error())
}


func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
