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
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/response"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// resultConfigurationType defines the type for the Configuration element in resultType
type resultConfigurationType map[string]interface{}

// resultType defines the result returned for a get configuration request.
type resultType struct {
	Configuration resultConfigurationType `json:"configuration"`
}

// GetExecutor defines a contract for getting configuration for a service.
type GetExecutor interface {
	Do(ctx context.Context, service string) (string, error)
}

// get contains references to dependencies required to execute a get configuration request.
type get struct {
	executor      GetExecutor
	loggingClient logger.LoggingClient
}

// New is a factory function that returns an initialized get struct.
func New(executor GetExecutor, loggingClient logger.LoggingClient) *get {
	return &get{
		executor:      executor,
		loggingClient: loggingClient,
	}
}

// Do fulfills the GetConfig contract and implements the retrieval of configuration for multiple services.
func (g get) Do(ctx context.Context, services []string) interface{} {
	result := resultType{
		Configuration: resultConfigurationType{},
	}
	for _, service := range services {
		c, err := g.executor.Do(ctx, service)
		if err != nil {
			g.loggingClient.Error(fmt.Sprintf(err.Error()))
			result.Configuration[service] = fmt.Sprintf(err.Error())
			continue
		}
		result.Configuration[service] = response.Process(c, g.loggingClient)
	}
	return result
}
