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
 *
 * @microservice: core-command-go service
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/edgexfoundry/edgex-go/core/command"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
)

const (
	configFile string = "res/configuration.json"
)

var configuration command.ConfigurationStruct

func main() {
	var loggingClient = logger.NewClient(command.SERVICENAME, "")
	// Load configuration data
	err := readConfigurationFile(configFile)
	if err != nil {
		loggingClient.Error("Could not read config file(" + configFile + "): " + err.Error())
		return
	}

	// Setup Logging
	loggingClient.RemoteUrl = configuration.LoggingRemoteURL
	loggingClient.LogFilePath = configuration.LogFile

	command.Start(configuration, loggingClient)
}

func readConfigurationFile(path string) error {
	// Read the configuration file
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Decode the configuration as JSON
	return json.Unmarshal(contents, &configuration)
}
