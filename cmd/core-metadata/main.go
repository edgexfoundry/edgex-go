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
	"io/ioutil"
	"os"

	"github.com/edgexfoundry/edgex-go/core/metadata"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
)

const (
	CONFIG = "res/configuration.json"
)

func main() {
	var loggingClient = logger.NewClient(metadata.METADATASERVICENAME, "")

	// Load configuration data
	configuration, err := readConfigurationFile(CONFIG)
	if err != nil {
		loggingClient.Error("Could not read configuration file(" + CONFIG + "): " + err.Error())
		os.Exit(1)
	}

	// Update logging based on configuration
	loggingClient.RemoteUrl = configuration.LoggingRemoteURL
	loggingClient.LogFilePath = configuration.LoggingFile

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
