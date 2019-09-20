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

package tokenconfig

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
type TokenConfFile map[string]ServiceKey

type ServiceKey struct {
	UseDefaults           bool                   `json:"edgex_use_defaults"` // BUG - change to underscores in man page
	CustomPolicy          map[string]interface{} `json:"custom_policy"`      // JSON serialization of HCL
	CustomTokenParameters map[string]interface{} `json:"custom_token_parameters"`
}

/* Note that above types must be exported in order to visible to JSON marshaller */
