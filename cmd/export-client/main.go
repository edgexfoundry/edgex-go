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
	"os"
	"os/signal"
	"syscall"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/export/client"
	"github.com/edgexfoundry/edgex-go/pkg/config"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	var (
		useConsul  = flag.String("consul", "", "Should the service use consul?")
		useProfile = flag.String("profile", "default", "Specify a profile other than default.")
	)
	flag.Parse()

	configuration := &client.ConfigurationStruct{}
	err := config.LoadFromFile(*useProfile, configuration)
	if err != nil {
		logger.Error(err.Error(), zap.String("version", edgex.Version))
		return
	}

	//Determine if configuration should be overridden from Consul
	var consulMsg string
	if *useConsul == "y" {
		consulMsg = "Loading configuration from Consul..."
		err := client.ConnectToConsul(*configuration)
		if err != nil {
			logger.Error(err.Error(), zap.String("version", edgex.Version))
			return //end program since user explicitly told us to use Consul.
		}
	} else {
		consulMsg = "Bypassing Consul configuration..."
	}

	logger.Info(consulMsg, zap.String("version", edgex.Version))

	err = client.Init(*configuration, logger)

	errs := make(chan error, 2)

	client.StartHTTPServer(*configuration, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	c := <-errs
	logger.Info("terminated", zap.String("error", c.Error()))
}
