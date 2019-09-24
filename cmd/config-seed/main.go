/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/seed/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {
	var configDir, profileDir string
	var dirCmd string
	var dirProperties string
	var overwriteConfig bool

	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")
	flag.StringVar(&dirProperties, "props", "./res/properties", "Specify alternate properties location as absolute path")
	flag.StringVar(&dirProperties, "r", "./res/properties", "Specify alternate properties location as absolute path")
	flag.StringVar(&dirCmd, "cmd", "../cmd", "Specify alternate cmd location as absolute path")
	flag.StringVar(&dirCmd, "c", "../cmd", "Specify alternate cmd location as absolute path")
	flag.BoolVar(&overwriteConfig, "overwrite", false, "Overwrite configuration in Registry")
	flag.BoolVar(&overwriteConfig, "o", false, "Overwrite configuration in Registry")

	flag.Usage = usage.HelpCallbackConfigSeed
	flag.Parse()

	bootstrap(configDir, profileDir)
	ok := config.Init()
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", clients.ConfigSeedServiceKey))
		os.Exit(1)
	}
	config.LoggingClient.Info("Service dependencies resolved...")

	err := config.ImportProperties(dirProperties)
	if err != nil {
		config.LoggingClient.Error(err.Error())
	}

	err = config.ImportConfiguration(dirCmd, profileDir, overwriteConfig)
	if err != nil {
		config.LoggingClient.Error(err.Error())
	}

	err = config.ImportSecurityConfiguration()
	if err != nil {
		config.LoggingClient.Error(err.Error())
	}

	os.Exit(0)
}

func bootstrap(configDir, profileDir string) {
	deps := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go config.Retry(configDir, profileDir, internal.BootTimeoutDefault, &wg, deps)
	go func(ch chan error) {
		for {
			select {
			case e, ok := <-ch:
				if ok {
					config.LoggingClient.Error(e.Error())
				} else {
					return
				}
			}
		}
	}(deps)

	wg.Wait()
}

func logBeforeInit(err error) {
	l := logger.NewClient(clients.ConfigSeedServiceKey, false, "", models.InfoLog)
	l.Error(err.Error())
}
