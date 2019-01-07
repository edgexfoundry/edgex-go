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
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/seed/config"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

func main() {
	var useProfile string
	var dirCmd string
	var dirProperties string
	var overwriteConfig bool

	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.StringVar(&dirProperties, "props", "./res/properties", "Specify alternate properties location as absolute path")
	flag.StringVar(&dirProperties, "r", "./res/properties", "Specify alternate properties location as absolute path")
	flag.StringVar(&dirCmd, "cmd", "../cmd", "Specify alternate cmd location as absolute path")
	flag.StringVar(&dirCmd, "c", "../cmd", "Specify alternate cmd location as absolute path")
	flag.BoolVar(&overwriteConfig, "overwrite", false, "Overwrite configuration in Consul")
	flag.BoolVar(&overwriteConfig, "o", false, "Overwrite configuration in Consul")

	flag.Usage = usage.HelpCallbackConfigSeed
	flag.Parse()

	bootstrap(useProfile)
	ok := config.Init()
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", internal.ConfigSeedServiceKey))
		return
	}
	config.LoggingClient.Info("Service dependencies resolved...")
	err := config.ImportProperties(dirProperties)
	if err != nil {
		config.LoggingClient.Error(err.Error())
	}
	err = config.ImportConfiguration(dirCmd, useProfile, overwriteConfig)
	if err != nil {
		config.LoggingClient.Error(err.Error())
	}
}

func bootstrap(profile string) {
	deps := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go config.Retry(profile, internal.BootTimeoutDefault, &wg, deps)
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
	l := logger.NewClient(internal.ConfigSeedServiceKey, false, "", logger.InfoLog)
	l.Error(err.Error())
}
