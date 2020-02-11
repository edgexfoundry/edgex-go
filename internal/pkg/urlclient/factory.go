/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

// urlclient provides functions to integrate the client code in go-mod-core-contracts with application specific code
package urlclient

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/retry"
)

// New is a factory function that uses parameters defined in edgex-go to decide which implementation of URLClient to use
func New(useRegistry bool, endpoint *endpoint.Endpoint, url string) interfaces.URLClient {
	if useRegistry {
		return retry.New(endpoint.Monitor(), 500, 10)
	}

	return local.New(url)
}
