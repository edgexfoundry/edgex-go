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
	"github.com/edgexfoundry/edgex-go/pkg/usage"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	logger.Info(fmt.Sprintf("Starting %s %s", client.ExportClient, edgex.Version))

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

	configuration := &client.ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		logger.Error(err.Error(), zap.String("version", edgex.Version))
		return
	}

	//Determine if configuration should be overridden from Consul
	var consulMsg string
	if useConsul {
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
