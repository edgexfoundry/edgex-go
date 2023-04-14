//
// Copyright (c) 2019-2023 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0
//

package fileprovider

/*

Example config file

{
  "service-name": {
    "edgex_use_defaults": true,
    "custom_policy": [
      {
        "path": {
          "secret/non/standard/location/*": {
            "capabilities": [ "list", "read" ]
          }
        }
      }
    ],
    "custom_token_parameters": { },
    "file_permissions": {
      "uid": 0,
      "gid": 0,
      "mode_octal": "0600"
	}
  }
}

*/

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/token/fileioperformer"
)

const (
	addSecretstoreTokensEnvKey = "EDGEX_ADD_SECRETSTORE_TOKENS" // nolint:gosec
)

type TokenConfFile map[string]ServiceKey

type FilePermissions struct {
	Uid       *int    `json:"uid,omitempty"`
	Gid       *int    `json:"gid,omitempty"`
	ModeOctal *string `json:"mode_octal,omitempty"`
}

type ServiceKey struct {
	UseDefaults     bool                   `json:"edgex_use_defaults"`
	CustomPolicy    map[string]interface{} `json:"custom_policy"` // JSON serialization of HCL
	FilePermissions *FilePermissions       `json:"file_permissions,omitempty"`
}

func LoadTokenConfig(fileOpener fileioperformer.FileIoPerformer, path string, tokenConf *TokenConfFile) error {
	reader, err := fileOpener.OpenFileReader(path, os.O_RDONLY, 0400)
	if err != nil {
		return err
	}
	readCloser := fileioperformer.MakeReadCloser(reader)
	defer readCloser.Close()

	err = json.NewDecoder(readCloser).Decode(tokenConf)
	if err != nil {
		return err
	}

	return nil
}

// GetTokenConfigFromEnv function gets a list of token service keys
// from environment variable and populates the default configuration with
// default token parameters and policies
// the function returns a TokenConfFile map instance and error if any
// if the environment variable is not present or the value of that is empty
// then it will return empty map
// if the value for the list is not well-formed, not comma-separated
// then it will return an error
func GetTokenConfigFromEnv() (TokenConfFile, error) {
	emptyTokenConfig := make(TokenConfFile)

	addTokenList := os.Getenv(addSecretstoreTokensEnvKey)
	if strings.TrimSpace(addTokenList) == "" {
		return emptyTokenConfig, nil
	}

	// the list of service names is comma-separated
	tokenConfigFromEnv := make(TokenConfFile)
	serviceNameList := strings.Split(addTokenList, ",")
	serviceNameRegx := regexp.MustCompile(secretstore.ServiceNameValidationRegx)

	for _, name := range serviceNameList {
		serviceName := strings.TrimSpace(name)
		if serviceName == "" {
			// skipping the empty name cases, ie. treating it as non-existent service
			continue
		}

		if !serviceNameRegx.MatchString(serviceName) {
			return emptyTokenConfig, fmt.Errorf("invalid service name: %s as key from environment variable", serviceName)
		}

		// with default service configuration
		tokenConfigFromEnv[serviceName] = ServiceKey{
			UseDefaults: true,
		}
	}

	return tokenConfigFromEnv, nil
}

// mergeWith function takes another TokenConfFile and merges with the current one
// to return a new TokenConfFile.
// The merging is based on the key of TokenConfFile's map
// if the key of another has already been existing on the current one,
// then the former (ie. from another) will replaces the latter (ie. from tf).
func (tf TokenConfFile) mergeWith(another TokenConfFile) TokenConfFile {
	if len(another) == 0 {
		// nothing to be merged, return itself
		return tf
	}

	if len(tf) == 0 {
		// itself is empty, just use another
		return another
	}

	// deep copy the tf first
	mergedMap := make(TokenConfFile)
	for k, v := range tf {
		mergedMap[k] = v
	}

	for key, value := range another {
		mergedMap[key] = value
	}

	return mergedMap
}

// keyExists function returns true if the input key exists in the TokenConfFile map
func (tf TokenConfFile) keyExists(key string) bool {
	_, exists := tf[key]
	return exists
}
