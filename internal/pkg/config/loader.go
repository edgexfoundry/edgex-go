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

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

const (
	configDirectory = "./res"

	configDirEnv = "EDGEX_CONF_DIR"
)

// TODO: Where is this set? It doesn't appear to be set from anywhere meaning if we fail to locate the config
//       file, the logging statements below will panic.
// deprecated
var LoggingClient logger.LoggingClient

// deprecated
func LoadFromFile(configDir, profileDir string, configuration interface{}) error {
	path := determinePath(configDir)
	fileName := path + "/" + internal.ConfigFileName //default profileDir
	if len(profileDir) > 0 {
		fileName = path + "/" + profileDir + "/" + internal.ConfigFileName
	}
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		msg := fmt.Sprintf("could not load configuration file (%s): %s", fileName, err.Error())
		LoggingClient.Error(msg)
		return errors.New(msg)
	}

	// Decode the configuration from TOML
	err = toml.Unmarshal(contents, configuration)
	if err != nil {
		msg := fmt.Sprintf("unable to parse configuration file (%s): %s", fileName, err.Error())
		LoggingClient.Error(msg)
		return errors.New(msg)
	}

	return nil
}

// deprecated
func determinePath(path string) string {
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
	files, _ := filepath.Glob("res/*/*.toml")
	files2, _ := filepath.Glob("res/configuration.toml")

	for _, x := range files2 {
		files = append(files, x)
	}

	if len(files) == 0 {
		return fmt.Errorf("There are no toml files")
	}

	for _, f := range files {
		profileDir := f[len("res") : len(f)-len("/configuration.toml")]
		if profileDir != "" {
			// remove the dash
			profileDir = profileDir[1:]
		}
		err := LoadFromFile("", profileDir, configuration)
		if err != nil {
			return fmt.Errorf("Error loading toml file %s: %v", profileDir, err)
		}
	}
	return nil
}
