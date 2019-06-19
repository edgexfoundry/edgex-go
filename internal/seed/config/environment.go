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
	"github.com/pelletier/go-toml"
	"os"
	"strings"
)

// environment is receiver that holds environment variables and encapsulates toml.Tree-based configuration field
// overrides.  Assumes "_" embedded in environment variable key separates substructs; e.g. foo_bar_baz might refer to
//
// 		type foo struct {
// 			bar struct {
//          	baz string
//  		}
//		}
//
// Two types of environment variable overrides are supported: generic and service-specific.  Service-specific
// overrides have a service name prefix (separated from the rest of the key name by an "_").
type environment struct {
	env map[string]interface{}
}

// NewEnvironment constructor reads/stores os.Environ() for use by OverrideFromEnvironment() method.
func NewEnvironment() *environment {
	osEnv := os.Environ()
	e := &environment{
		env: make(map[string]interface{}, len(osEnv)),
	}
	for _, env := range osEnv {
		kv := strings.Split(env, "=")
		if len(kv) == 2 && len(kv[0]) > 0 && len(kv[1]) > 0 {
			e.env[strings.Replace(kv[0], "_", ".", -1)] = kv[1]
		}
	}
	return e
}

// OverrideFromEnvironment method replaces values in the toml.Tree for matching environment variable keys.
func (e *environment) OverrideFromEnvironment(serviceName string, tree *toml.Tree) *toml.Tree {
	serviceName += "."
	for k, v := range e.env {
		switch {
		case tree.Has(k):
			// global key
			tree.Set(k, v)
		case strings.HasPrefix(k, serviceName):
			// service-specific key
			tree.Set(strings.TrimPrefix(k, serviceName), v)
		default:
			// do nothing
		}
	}
	return tree
}
