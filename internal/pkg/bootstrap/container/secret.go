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

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/go-mod-secrets/pkg"
)

// SecretClientName contains the name of the registry.Client implementation in the DIC.
var SecretClientName = di.TypeInstanceToName(pkg.SecretClient{})

// SecretClientFrom helper function queries the DIC and returns the pkg.SecretClient implementation.
func SecretClientFrom(get di.Get) *pkg.SecretClient {
	return get(SecretClientName).(*pkg.SecretClient)
}
