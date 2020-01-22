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
	"github.com/edgexfoundry/edgex-go/internal/seed/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {
	var configProviderUrl string
	var configDir, profileDir string
	var dirCmd string
	var dirProperties string
	var overwriteConfig bool

	flag.StringVar(&configProviderUrl, "configProvider", "", "Indicates to use Configuration Provider service at specified URL. Format: {type}.{protocol}://{host}:{port} ex: consul.http://localhost:8500.")
	flag.StringVar(&configProviderUrl, "cp", "", "Indicates to use Configuration Provider service at specified URL. Format: {type}.{protocol}://{host}:{port} ex: consul.http://localhost:8500.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")
	flag.StringVar(&dirProperties, "props", "./res/properties", "Specify alternate properties location as absolute path")
	flag.StringVar(&dirProperties, "r", "./res/properties", "Specify alternate properties location as absolute path")
	flag.StringVar(&dirCmd, "cmd", "../cmd", "Specify alternate cmd location as absolute path")
	flag.StringVar(&dirCmd, "c", "../cmd", "Specify alternate cmd location as absolute path")
	flag.BoolVar(&overwriteConfig, "overwrite", false, "Overwrite configuration in Config Service")
	flag.BoolVar(&overwriteConfig, "o", false, "Overwrite configuration in Config Service")

	flag.Usage = helpCallback
	flag.Parse()

	providerOverride, exists := os.LookupEnv(internal.ConfigProviderEnvVar)
	if exists {
		configProviderUrl = providerOverride
	}

	bootstrap(configDir, profileDir, configProviderUrl)
	ok := config.Init()
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", clients.ConfigSeedServiceKey))
		os.Exit(1)
	}
	config.LoggingClient.Info("Service dependencies resolved...")

	err := config.ImportProperties(dirProperties, configProviderUrl)
	if err != nil {
		config.LoggingClient.Error(err.Error())
	}

	err = config.ImportConfiguration(dirCmd, profileDir, configProviderUrl, overwriteConfig)
	if err != nil {
		config.LoggingClient.Error(err.Error())
	}

	err = config.ImportSecurityConfiguration(configProviderUrl)
	if err != nil {
		config.LoggingClient.Error(err.Error())
	}

	os.Exit(0)
}

func bootstrap(configDir, profileDir string, configProviderUrl string) {
	deps := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go config.Retry(configDir, profileDir, configProviderUrl, internal.BootTimeoutDefault, &wg, deps)
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

func helpCallback() {
	fmt.Printf(
		"Usage: %s [options]\n"+
			"Server Options:\n"+
			"   -cp, --configProvider           Indicates to use Configuration Provider service at specified URL.\n"+
			"                                   URL Format: {type}.{protocol}://{host}:{port} ex: consul.http://localhost:8500\n"+
			"   -c, --cmd <dir>                 Provide absolute path to \"cmd\" directory containing EdgeX service configuration\n"+
			"   -o, --overwrite                 Indicates service should overwrite any entries already present in the configuration\n"+
			"   -p, --profile <name>            Indicate configuration profile other than default\n"+
			"   -r, --props <dir>               Provide alternate location for legacy application.properties files\n"+
			"   --confdir                       Specify local configuration directory\n"+
			"\n"+
			"Common Options:\n"+
			"   -h, --help                      Show this message\n",
		os.Args[0])
	os.Exit(0)
}
