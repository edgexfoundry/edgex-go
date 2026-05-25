//
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2025 IOTech Ltd
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

// AuthTokenLoader returns authorization token for secret store
type AuthTokenLoader interface {
	// Load loads and returns authorization token
	Load(path string) (string, error)
	// ReadEntityId reads the token file and returns the entity id
	ReadEntityId(path string) (string, error)
}
