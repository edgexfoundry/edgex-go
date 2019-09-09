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

package config

import (
	"github.com/magiconair/properties/assert"
	"github.com/pelletier/go-toml"
	"os"
	"testing"
)

const (
	serviceName         = "serviceName"
	serviceNameHyphen   = "service-Name"
	serviceNameUndScore = "service_Name"
	envValue            = "envValue"
	rootKey             = "rootKey"
	rootValue           = "rootValue"
	sub                 = "sub"
	subKey              = "subKey"
	subValue            = "subValue"
)

const testToml = `
` + rootKey + `="` + rootValue + `"
[` + sub + `]
` + subKey + `="` + subValue + `"`

func newSUT(t *testing.T, envKey string, envValue string) (*toml.Tree, *environment) {
	os.Clearenv()
	if err := os.Setenv(envKey, envValue); err != nil {
		t.Fail()
	}

	tree, err := toml.Load(testToml)
	if err != nil {
		t.Fail()
	}
	return tree, NewEnvironment()
}

func TestKeyMatchOverwritesValue(t *testing.T) {
	var tests = []struct {
		name          string
		key           string
		envKey        string
		envValue      string
		serviceName   string
		expectedValue string
	}{
		{"generic root", rootKey, rootKey, envValue, serviceName, envValue},
		{"generic sub", sub + "." + subKey, sub + "." + subKey, envValue, serviceName, envValue},
		{"service root", rootKey, serviceName + "." + rootKey, envValue, serviceName, envValue},
		{"service sub", sub + "." + subKey, serviceName + "." + sub + "." + subKey, envValue, serviceName, envValue},
		{"service hyphen", rootKey, serviceNameHyphen + "." + rootKey, envValue, serviceNameHyphen, envValue},
		{"service hyphen sub", sub + "." + subKey, serviceNameHyphen + "." + sub + "." + subKey, envValue, serviceNameHyphen, envValue},
		{"service underscore", rootKey, serviceNameUndScore + "." + rootKey, envValue, serviceNameHyphen, envValue},
		{"service underscore sub", sub + "." + subKey, serviceNameUndScore + "." + sub + "." + subKey, envValue, serviceNameHyphen, envValue},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree, sut := newSUT(t, test.envKey, test.envValue)

			result := sut.OverrideFromEnvironment(test.serviceName, tree)

			assert.Equal(t, result.Get(test.key), test.expectedValue)
		})
	}
}

func TestNonMatchingKeyDoesNotOverwritesValue(t *testing.T) {
	var tests = []struct {
		name          string
		key           string
		envKey        string
		envValue      string
		serviceName   string
		expectedValue string
	}{
		{"root", rootKey, serviceName + "." + rootKey, envValue, serviceName + "_other", rootValue},
		{"sub", sub + "." + subKey, serviceName + "." + sub + "." + subKey, envValue, serviceName + "_other", subValue},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree, sut := newSUT(t, test.envKey, test.envValue)

			result := sut.OverrideFromEnvironment(test.serviceName, tree)

			assert.Equal(t, result.Get(test.key), test.expectedValue)
		})
	}
}
