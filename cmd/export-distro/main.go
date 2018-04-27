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
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/export/distro"
	"github.com/edgexfoundry/edgex-go/pkg/config"
	"github.com/edgexfoundry/edgex-go/pkg/usage"

	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting edgex export distro", zap.String("version", edgex.Version))

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
		logger.Error(err.Error(), zap.String("version", edgex.Version))
		return
	}

	//Determine if configuration should be overridden from Consul
	var consulMsg string
	if useConsul {
		consulMsg = "Loading configuration from Consul..."
		err := distro.ConnectToConsul(*configuration)
		if err != nil {
			logger.Error(err.Error(), zap.String("version", edgex.Version))
			return //end program since user explicitly told us to use Consul.
		}
	} else {
		consulMsg = "Bypassing Consul configuration..."
	}

	logger.Info(consulMsg, zap.String("version", edgex.Version))

	err = distro.Init(*configuration, logger)

	logger.Info("Starting distro")
	errs := make(chan error, 2)
	eventCh := make(chan *models.Event, 10)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// There can be another receivers that can be initialiced here
	distro.ZeroMQReceiver(eventCh)

	distro.Loop(errs, eventCh)

	logger.Info("terminated")
}
