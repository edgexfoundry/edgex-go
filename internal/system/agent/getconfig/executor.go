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

package getconfig

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
)

// clientFactory defines contract for creating/retrieving a general client.
type clientFactory interface {
	New(serviceName string) (general.GeneralClient, error)
}

// executor contains references to dependencies required to execute a get configuration request.
type executor struct {
	clientFactory clientFactory
}

// NewExecutor is a factory function that returns an initialized executor struct.
func NewExecutor(clientFactory clientFactory) *executor {
	return &executor{
		clientFactory: clientFactory,
	}
}

// Do fulfills the GetExecutor contract and implements the functionality to retrieve a service's configuration.
func (e executor) Do(ctx context.Context, serviceName string) (string, error) {
	client, err := e.clientFactory.New(serviceName)
	if err != nil {
		return "", err
	}
	return client.FetchConfiguration(ctx)
}
