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
	"fmt"
)

const (
	configFile string = "res/configuration.json"
)

var loggingClient logger.LoggingClient

func main() {

	// Load configuration data
	configuration, err := readConfigurationFile(configFile)
	if err != nil {
		loggingClient = logger.NewClient(command.COMMANDSERVICENAME, false, "")
		loggingClient.Error("Could not load configuration (" + configFile + "): " + err.Error())
		return
	}

	// Setup Logging
	logTarget := setLoggingTarget(*configuration)
	var loggingClient = logger.NewClient(configuration.ApplicationName, configuration.EnableRemoteLogging, logTarget)

	command.Start(*configuration, loggingClient)
}

func readConfigurationFile(path string) (*command.ConfigurationStruct, error) {
	var configuration command.ConfigurationStruct
	// Read the configuration file
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading configuration file: " + err.Error())
		return nil, err
	}

	// Decode the configuration as JSON
	err = json.Unmarshal(contents, &configuration)
	if err != nil {
		fmt.Println("Error parsing configuration file: " + err.Error())
		return nil, err
	}

	return &configuration, nil
}

func setLoggingTarget(conf command.ConfigurationStruct) string {
	logTarget := conf.LoggingRemoteURL
	if !conf.EnableRemoteLogging {
		return conf.LogFile
	}
	return logTarget
}
