/*******************************************************************************
 * Copyright 2022 Intel Corp.
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
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/runtimetokenprovider"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// RuntimeTokenProviderInterfaceName contains the name of the runtimetokenprovider.RuntimeTokenProvider implementation in the DIC.
var RuntimeTokenProviderInterfaceName = di.TypeInstanceToName((*runtimetokenprovider.RuntimeTokenProvider)(nil))

// RuntimeTokenProviderFrom helper function queries the DIC and returns the runtimetokenprovider.RuntimeTokenProvider implementation.
func RuntimeTokenProviderFrom(get di.Get) runtimetokenprovider.RuntimeTokenProvider {
	runtimeTokenProvider, ok := get(RuntimeTokenProviderInterfaceName).(runtimetokenprovider.RuntimeTokenProvider)
	if !ok {
		return nil
	}

	return runtimeTokenProvider
}
