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
 * @microservice: core-metadata-go service
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tsconn23/edgex-go"
	"github.com/tsconn23/edgex-go/core/metadata"
	logger "github.com/tsconn23/edgex-go/support/logging-client"
)

const (
	configFile = "res/configuration.json"
)

var loggingClient logger.LoggingClient

func main() {
	// Load configuration data
	configuration, err := readConfigurationFile(configFile)
	if err != nil {
		loggingClient = logger.NewClient(metadata.METADATASERVICENAME, false, "")
		loggingClient.Error("Could not load configuration (" + configFile + "): " + err.Error())
		return
	}

	logTarget := setLoggingTarget(*configuration)
	// Create Logger (Default Parameters)
	loggingClient = logger.NewClient(configuration.ApplicationName, configuration.EnableRemoteLogging, logTarget)
	loggingClient.Info("Starting core-metadata " + edgex.Version)

	loggingClient.Info(fmt.Sprintf("Starting %s %s ", metadata.METADATASERVICENAME, edgex.Version))

	metadata.Start(*configuration, loggingClient)
}

// Read the configuration file and
func readConfigurationFile(path string) (*metadata.ConfigurationStruct, error) {
	var configuration metadata.ConfigurationStruct
	// Read the configuration file
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode the configuration as JSON
	err = json.Unmarshal(contents, &configuration)
	if err != nil {
		return nil, err
	}

	return &configuration, nil
}

func setLoggingTarget(conf metadata.ConfigurationStruct) string {
	logTarget := conf.LoggingRemoteURL
	if !conf.EnableRemoteLogging {
		return conf.LoggingFile
	}
	return logTarget
}
