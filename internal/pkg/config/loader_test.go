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
	"os"
	"testing"
)

const testProfile = "test"

type TestConfigurationStruct struct {
	ApplicationName string
}

func TestLoadFromDefault(t *testing.T) {
	configuration := &TestConfigurationStruct{}
	err := LoadFromFile(testProfile, configuration)
	if err != nil {
		t.Error(err)
	}

	if len(configuration.ApplicationName) == 0 {
		t.Errorf("configuration.ApplicationName is zero length.")
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	configuration := &TestConfigurationStruct{}
	os.Setenv(configDirEnv, configDirectory)
	err := LoadFromFile(testProfile, configuration)
	if err != nil {
		t.Error(err)
	}

	if len(configuration.ApplicationName) == 0 {
		t.Errorf("configuration.ApplicationName is zero length.")
	}
}

func TestVerify(t *testing.T) {
	configuration := &TestConfigurationStruct{}
	err := VerifyTomlFiles(configuration)
	if err != nil {
		t.Errorf("Error parsing sample configuration files: %v", err)
	}
}

func TestDetermineConfigFile(t *testing.T) {
	var tests = []struct {
		profile string
		file    string
	}{
		{"", configDefault},
		{"go", "configuration-go.toml"},
		{"another", "configuration-another.toml"},
	}
	for _, tt := range tests {
		t.Run(tt.profile, func(t *testing.T) {
			file := determineConfigFile(tt.profile)
			if tt.file != file {
				t.Errorf("filename for profile %s should be %s, not %s", tt.profile, tt.file, file)
			}
		})
	}

}
