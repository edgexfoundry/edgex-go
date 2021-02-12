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

import (
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	secretstoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
)

// TokenProvider is the interface that the main program expects
// for implemeneting token generation
type TokenProvider interface {
	// Set configuration
	SetConfiguration(secretStore secretstoreConfig.SecretStoreInfo, tokenConfig config.TokenFileProviderInfo)
	// Generate tokens
	Run() error
}
