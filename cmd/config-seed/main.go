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
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/seed/config"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"sync"
)

var bootTimeout int = 30000 //Once we start the V2 configuration rework, this will be config driven

func main() {
	var useProfile string
	var dirProperties string

	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.StringVar(&dirProperties, "props", "./res/properties", "Specify alternate properties location")
	flag.StringVar(&dirProperties, "r", "./res/properties", "Specify alternate properties location")

	flag.Usage = usage.HelpCallback
	flag.Parse()

	bootstrap(useProfile)
	ok := config.Init()
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", internal.ConfigSeedServiceKey))
		return
	}
	config.LoggingClient.Info("Service dependencies resolved...")
	config.ImportProperties(dirProperties)
}

func bootstrap(profile string) {
	deps := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go config.Retry(profile, bootTimeout, &wg, deps)
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
	l := logger.NewClient(internal.ConfigSeedServiceKey, false, "")
	l.Error(err.Error())
}
