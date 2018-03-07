/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/tsconn23/edgex-go"
	"github.com/tsconn23/edgex-go/core/data"
	"github.com/tsconn23/edgex-go/support/logging-client"
)

const (
	configFile string = "./res/configuration.json"
)

var loggingClient logger.LoggingClient

func main() {
	start := time.Now()

	// Load configuration data
	configuration, err := readConfigurationFile(configFile)
	if err != nil {
		loggingClient = logger.NewClient(data.COREDATASERVICENAME, false, "")
		loggingClient.Error("Could not load configuration (" + configFile + "): " + err.Error())
		return
	}

	logTarget := setLoggingTarget(*configuration)
	// Create Logger (Default Parameters)
	loggingClient = logger.NewClient(configuration.Applicationname, configuration.EnableRemoteLogging, logTarget)

	loggingClient.Info(fmt.Sprintf("Starting %s %s ", data.COREDATASERVICENAME, edgex.Version))

	data.Init(*configuration, loggingClient)

	r := data.LoadRestRoutes()
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(5000), "Request timed out")
	loggingClient.Info(configuration.Appopenmsg, "")

	// Time it took to start service
	loggingClient.Info("Service started in: "+time.Since(start).String(), "")
	loggingClient.Info("Listening on port: " + strconv.Itoa(configuration.Serverport))

	loggingClient.Error(http.ListenAndServe(":"+strconv.Itoa(configuration.Serverport), r).Error())
}

// Read the configuration file and update configuration struct
func readConfigurationFile(path string) (*data.ConfigurationStruct, error) {
	var configuration data.ConfigurationStruct
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

func setLoggingTarget(conf data.ConfigurationStruct) string {
	logTarget := conf.Loggingremoteurl
	if !conf.EnableRemoteLogging {
		return conf.Loggingfile
	}
	return logTarget
}
