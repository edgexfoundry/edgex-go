/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright 2018 Joan Duran
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

package general

import (
	"context"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

type GeneralClient interface {
	FetchConfiguration(ctx context.Context) (string, error)
	FetchMetrics(ctx context.Context) (string, error)
}

type generalRestClient struct {
	url      string
	endpoint clients.Endpointer
}

func NewGeneralClient(params types.EndpointParams, m clients.Endpointer) GeneralClient {
	gc := generalRestClient{endpoint: m}
	gc.init(params)
	return &gc
}

func (gc *generalRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go gc.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					gc.url = url
				}
			}
		}(ch)
	} else {
		gc.url = params.Url
	}
}

// FetchConfiguration fetch configuration information from the service.
func (gc *generalRestClient) FetchConfiguration(ctx context.Context) (string, error) {
	body, err := clients.GetRequest(gc.url+clients.ApiConfigRoute, ctx)
	return string(body), err
}

// FetchMetrics fetch metrics information from the service.
func (gc *generalRestClient) FetchMetrics(ctx context.Context) (string, error) {
	body, err := clients.GetRequest(gc.url+clients.ApiMetricsRoute, ctx)
	return string(body), err
}
