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

package setconfig

import (
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"
)

// resultConfigurationType defines the type for the Configuration element in resultType
type resultConfigurationType map[string]responses.SetConfigResponse

// resultType defines the result returned for a set configuration request.
type resultType struct {
	Configuration resultConfigurationType `json:"configuration"`
}

// GetExecutor defines a contract for setting a service's configuration.
type SetExecutor interface {
	Do(service string, sc requests.SetConfigRequest) responses.SetConfigResponse
}

// set contains references to dependencies required to execute a set configuration request.
type set struct {
	executor SetExecutor
}

// New is a factory function that returns an initialized set struct.
func New(executor SetExecutor) *set {
	return &set{
		executor: executor,
	}
}

// Do fulfills the SetConfig contract and implements the setting of configuration for multiple services.
func (s set) Do(services []string, sc requests.SetConfigRequest) interface{} {
	result := resultType{
		Configuration: resultConfigurationType{},
	}

	// Loop over services and accumulate the response (i.e. "result") to return to requester.
	for _, service := range services {
		result.Configuration[service] = s.executor.Do(service, sc)
	}
	return result
}
