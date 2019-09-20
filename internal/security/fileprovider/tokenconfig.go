//
// Copyright (c) 2019 Intel Corporation
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
// SPDX-License-Identifier: Apache-2.0'
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
    "custom_token_parameters": { }
  }
}

*/

import (
	"encoding/json"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/fileioperformer"
)

type TokenConfFile map[string]ServiceKey

type ServiceKey struct {
	UseDefaults           bool                   `json:"edgex_use_defaults"`
	CustomPolicy          map[string]interface{} `json:"custom_policy"` // JSON serialization of HCL
	CustomTokenParameters map[string]interface{} `json:"custom_token_parameters"`
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
