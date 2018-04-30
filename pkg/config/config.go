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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	configDirectory = "./res"
	configDefault   = "configuration.toml"

	configDirEnv = "EDGEX_CONF_DIR"
)

var confDir = flag.String("confdir", "", "Specify local configuration directory")

func LoadFromFile(profile string, configuration interface{}) error {
	path := determinePath()
	fileName := path + "/" + determineConfigFile(profile)

	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("could not load configuration file (%s): %v", fileName, err.Error())
	}

	// Decode the configuration from TOML
	err = toml.Unmarshal(contents, configuration)
	if err != nil {
		return fmt.Errorf("unable to parse configuration file (%s): %v", fileName, err.Error())
	}

	return nil
}

func determineConfigFile(profile string) string {
	if profile == "" {
		return configDefault
	}
	return "configuration-" + profile + ".toml"
}

func determinePath() string {
	flag.Parse()

	path := *confDir

	if len(path) == 0 { //No cmd line param passed
		//Assumption: one service per container means only one var is needed, set accordingly for each deployment.
		//For local dev, do not set this variable since configs are all named the same.
		path = os.Getenv(configDirEnv)
	}

	if len(path) == 0 { //Var is not set
		path = configDirectory
	}

	return path
}

func VerifyTomlFiles(configuration interface{}) error {
	files, err := filepath.Glob("res/configuration*.toml")
	if err != nil {
		return fmt.Errorf("There are no toml files")
	}

	for _, f := range files {
		profile := f[len("res/configuration") : len(f)-len(".toml")]
		if profile != "" {
			// remove the dash
			profile = profile[1:]
		}
		err := LoadFromFile(profile, configuration)
		if err != nil {
			return fmt.Errorf("Error loading toml file %s: %v", profile, err)
		}
	}
	return nil
}
