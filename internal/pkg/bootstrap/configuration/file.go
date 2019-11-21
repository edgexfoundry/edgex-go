/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package configuration

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"

	"github.com/BurntSushi/toml"
)

// LoadFromFile attempts to read and unmarshal toml-based configuration into a configuration struct.
func LoadFromFile(configDir, profileDir, configFileName string, config interfaces.Configuration) error {
	// ported from determinePath() in internal/pkg/config/loader.go
	if len(configDir) == 0 {
		configDir = os.Getenv("EDGEX_CONF_DIR")
	}
	if len(configDir) == 0 {
		configDir = "./res"
	}

	// remainder is simplification of LoadFromFile() in internal/pkg/config/loader.go
	if len(profileDir) > 0 {
		profileDir += "/"
	}

	fileName := configDir + "/" + profileDir + configFileName

	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("could not load configuration file (%s): %s", fileName, err.Error())
	}
	if err = toml.Unmarshal(contents, config); err != nil {
		return fmt.Errorf("could not load configuration file (%s): %s", fileName, err.Error())
	}
	return nil
}
