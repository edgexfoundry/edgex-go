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
 *
 * @author: Tingyu Zeng, Dell
 * @version: 1.1.0
 *******************************************************************************/
package proxy

import (
	"testing"
)

func TestLoadTomlConfig(t *testing.T) {
	path := "testdata/tomltest.toml"
	config, err := LoadTomlConfig(path)
	if err != nil {
		t.Errorf("failed to parse toml file")
		t.Errorf(err.Error())
	}
	if config.SecretService.TokenPath != "/testdata/test-resp-init.json" {
		t.Errorf("failed to get correct value for tokenpath in the toml config file")
	}
	if config.EdgexServices["test"].Name != "test" {
		t.Errorf("failed to get correct name for test service in the toml config file")
	}

}
