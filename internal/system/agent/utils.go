/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
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
 *******************************************************************************/

package agent

import (
	"net/http"
	"flag"
	"io/ioutil"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"encoding/json"
)

// Test if the service is working
func pingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	_, err := w.Write([]byte("pong"))
	if err != nil {
		LoggingClient.Error("Error writing pong: " + err.Error())
	}
}

const (
	manifestDirectory = "./res"
	manifestDefault   = "manifest.toml"

	manifestDirEnv = "EDGEX_CONF_DIR"
)

var manifestDir = flag.String("manifestdir", "", "Specify local manifest directory")

func LoadFromFile(profile string, man interface{}) error {
	path := determinePath()
	fileName := path + "/" + determineManifestFile(profile)

	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("could not load configuration file (%s): %v", fileName, err.Error())
	}

	// Decode the configuration from TOML
	err = toml.Unmarshal(contents, man)
	if err != nil {
		return fmt.Errorf("unable to parse configuration file (%s): %v", fileName, err.Error())
	}

	return nil
}

func determineManifestFile(profile string) string {
	if profile == "" {
		return manifestDefault
	}
	return "manifest.toml"
}

func determinePath() string {
	flag.Parse()

	path := *manifestDir

	if len(path) == 0 { //No cmd line param passed
		// Assumption: one service per container means only one var is needed, set accordingly for each deployment.
		// For local dev, do not set this variable since manifests are all named the same.
		path = os.Getenv(manifestDirEnv)
	}

	if len(path) == 0 { //Var is not set
		path = manifestDirectory
	}

	return path
}
func ProcessResponse(response string) map[string]interface{} {
	rsp := make(map[string]interface{})
	err := json.Unmarshal([]byte(response), &rsp)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("ERROR: {%v}", err))
	}
	return rsp
}
