/*******************************************************************************
 * Copyright 2020 Redis Labs
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
 *
 * @author: Diana Atanasova
 * @author: Andre Srinivasan
 *******************************************************************************/
package secretread

import (
	"flag"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func LoadConfig(lc logger.LoggingClient) (Configuration, error) {
	lc.Info("loading configuration considering the secret store")
	configFileLocation := defineConfigFileLocation(lc)
	secureConfig := Configuration{}

	_, err := toml.DecodeFile(configFileLocation, &secureConfig)
	if err != nil {
		return Configuration{}, err
	}

	return secureConfig, err
}

func defineConfigFileLocation(lc logger.LoggingClient) string {
	var confdir string
	var profile string

	flag.StringVar(&profile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profile, "p", "", "Specify a profile other than default.")
	flag.StringVar(&confdir, "confdir", "", "configuration file")
	flag.Parse()

	directory := Confdir
	if len(confdir) > 0 {
		directory = confdir
	}

	configFile := directory + "/" + ConfigFileName
	if len(profile) > 0 {
		configFile = strings.Join([]string{directory, profile, ConfigFileName}, "/")
	}
	lc.Info(fmt.Sprintf("config file location: %s", configFile))
	return configFile
}
