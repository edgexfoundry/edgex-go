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
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/tsconn23/edgex-go"
	"github.com/tsconn23/edgex-go/core/metadata"
	"github.com/tsconn23/edgex-go/pkg/config"
	"github.com/tsconn23/edgex-go/pkg/heartbeat"
	logger "github.com/tsconn23/edgex-go/support/logging-client"
)

const (
	configFile = "res/configuration.json"
)

var loggingClient logger.LoggingClient

func main() {
	start := time.Now()
	var (
		useConsul = flag.String("consul", "", "Should the service use consul?")
		useProfile = flag.String("profile", "default", "Specify a profile other than default.")
	)
	flag.Parse()

	configuration := &metadata.ConfigurationStruct{}
	err := config.LoadFromFile(*useProfile, configuration)
	if err != nil {
		logBeforeTermination(err)
		return
	}

	logTarget := setLoggingTarget(*configuration)
	// Create Logger (Default Parameters)
	loggingClient = logger.NewClient(configuration.ApplicationName, configuration.EnableRemoteLogging, logTarget)
	loggingClient.Info("Starting core-metadata " + edgex.Version)

	loggingClient.Info(fmt.Sprintf("Starting %s %s ", metadata.METADATASERVICENAME, edgex.Version))

	heartbeat.Start(configuration.HeartBeatMsg, configuration.HeartBeatTime, loggingClient)

	metadata.Init(*configuration, loggingClient)
}

func logBeforeTermination(err error) {
	loggingClient = logger.NewClient(metadata.METADATASERVICENAME, false, "")
	loggingClient.Error(err.Error())
}


func setLoggingTarget(conf metadata.ConfigurationStruct) string {
	logTarget := conf.LoggingRemoteURL
	if !conf.EnableRemoteLogging {
		return conf.LoggingFile
	}
	return logTarget
}
