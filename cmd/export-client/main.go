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

	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export/client"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
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

	configuration := &client.ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		logBeforeInit(fmt.Errorf("%s: version: %s: err: %s", internal.ExportClientServiceKey, edgex.Version, err.Error()))
		return
	}

	//Determine if configuration should be overridden from Consul
	if useConsul {
		err := client.ConnectToConsul(*configuration)
		if err != nil {
			logBeforeInit(fmt.Errorf("%s: version: %s: err: %s", internal.ExportClientServiceKey, edgex.Version, err.Error()))
			return //end program since user explicitly told us to use Consul.
		}
	}

	err = client.Init(*configuration)
	if err != nil {
		logBeforeInit(fmt.Errorf("%s: could not initialize export client: %s", internal.ExportClientServiceKey, err.Error()))
		return
	}

	errs := make(chan error, 2)

	client.StartHTTPServer(*configuration, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	c := <-errs

	client.Destroy()

	client.LoggingClient.Error(fmt.Sprintf("%s: terminated with error(s): %s", internal.ExportClientServiceKey, c.Error()))
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.ExportClientServiceKey, false, "")
	l.Error(err.Error())
}
