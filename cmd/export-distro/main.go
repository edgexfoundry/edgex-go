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
	"os"
	"os/signal"
	"syscall"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export/distro"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func main() {
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

	configuration := &distro.ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		logBeforeInit(fmt.Errorf("%s: version: %s: err: %s", internal.ExportDistroServiceKey, edgex.Version, err.Error()))
		return
	}

	//Determine if configuration should be overridden from Consul
	if useConsul {
		err := distro.ConnectToConsul(*configuration)
		if err != nil {
			logBeforeInit(fmt.Errorf("%s: version: %s: err: %s", internal.ExportDistroServiceKey, edgex.Version, err.Error()))
			return //end program since user explicitly told us to use Consul.
		}
	}

	err = distro.Init(*configuration, useConsul)

	errs := make(chan error, 2)
	eventCh := make(chan *models.Event, 10)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// There can be another receivers that can be initialized here
	distro.ZeroMQReceiver(eventCh)

	distro.Loop(errs, eventCh)

	distro.LoggingClient.Info(fmt.Sprintf("%s: terminated", internal.ExportDistroServiceKey))
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.ExportDistroServiceKey, false, "")
	l.Error(err.Error())
}
