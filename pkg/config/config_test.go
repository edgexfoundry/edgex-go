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
	"testing"
	"os"
)

type TestConfigurationStruct struct {
	ApplicationName	string
}

func TestLoadFromDefault(t *testing.T) {
	configuration := &TestConfigurationStruct{}
	err := LoadFromFile("unit-test", configuration)
	if err != nil {
		t.Error(err)
	}

	if len(configuration.ApplicationName) == 0 {
		t.Errorf("configuration.ApplicationName is zero length.")
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	configuration := &TestConfigurationStruct{}
	os.Setenv("EDGEX_CONF_DIR", configDirectory)
	err := LoadFromFile("unit-test", configuration)
	if err != nil {
		t.Error(err)
	}

	if len(configuration.ApplicationName) == 0 {
		t.Errorf("configuration.ApplicationName is zero length.")
	}
}
