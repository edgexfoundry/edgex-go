/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
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
package endpoint

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/interfaces"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

type Endpoint struct {
	ctx            context.Context
	wg             *sync.WaitGroup
	RegistryClient *registry.Client
	serviceKey     string // The key of the service as found in the service registry (e.g. Consul)
	path           string // The path to the service's endpoint following port number in the URL
	interval       int    // The interval in milliseconds governing how often the client polls to keep the endpoint current

}

func New(
	ctx context.Context,
	wg *sync.WaitGroup,
	registryClient *registry.Client,
	serviceKey string,
	path string,
	interval int) *Endpoint {

	return &Endpoint{
		ctx:            ctx,
		wg:             wg,
		RegistryClient: registryClient,
		serviceKey:     serviceKey,
		path:           path,
		interval:       interval,
	}
}

func (e Endpoint) Monitor() chan interfaces.URLStream {
	ch := make(chan interfaces.URLStream, 1)
	url, err := e.buildURL()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout, err.Error())
	}
	ch <- interfaces.URLStream(url)

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		for {
			url, err := e.buildURL()
			if err != nil {
				_, _ = fmt.Fprintln(os.Stdout, err.Error())
			}
			ch <- interfaces.URLStream(url)
			time.Sleep(time.Millisecond * time.Duration(e.interval))

			// use ctx to drop out of infinite when ctx indicates done().
		}
	}()
	return ch
}

func (e Endpoint) buildURL() (string, error) {
	if e.RegistryClient != nil {
		endpoint, err := (*e.RegistryClient).GetServiceEndpoint(e.serviceKey)
		if err != nil {
			return "", fmt.Errorf("unable to get Service endpoint for %s: %s", e.serviceKey, err.Error())
		}
		return fmt.Sprintf("http://%s:%v%s", endpoint.Host, endpoint.Port, e.path), nil
	} else {
		return "", fmt.Errorf("unable to get Service endpoint for %s: Registry client is nil", e.serviceKey)
	}
}
