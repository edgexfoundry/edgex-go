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
 * @author: Trevor Conn, Dell
 * @version: 0.5.0
 *******************************************************************************/

package config

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

const (
	configDefault = "./res/configuration.json"
	configDocker = "./res/configuration-docker.json"
	configUnitTest = "./res/configuration-test.json"
)

func LoadFromFile(profile string, configuration interface{}) error {
	path := determineConfigFile(profile)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not load configuration file (%s): %v", path, err.Error())
	}

	// Decode the configuration from JSON
	err = json.Unmarshal(contents, configuration)
	if err != nil {
		return fmt.Errorf("unable to parse configuration file (%s): %v", path, err.Error())
	}

	return nil
}

func determineConfigFile(profile string) string {
	switch profile {
	case "docker":
		return configDocker
	case "unit-test":
		return configUnitTest
	default:
		return configDefault
	}
}
