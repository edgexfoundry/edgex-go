//
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2024-2025 IOTech Ltd
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

package authtokenloader

type secretStoreTokenFile struct {
	// Auth comes from the create token API
	Auth authObject `json:"auth"`
	// RootToken comes from the secret store-init response
	RootToken string `json:"root_token"`
}

type authObject struct {
	ClientToken string `json:"client_token"`
	EntityId    string `json:"entity_id"`
}
