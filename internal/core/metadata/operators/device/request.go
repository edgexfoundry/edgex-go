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

package device

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type RequestType int

const (
	Mock RequestType = 0
	Http RequestType = 1
)

type Requester interface {
	// Execute will invoke the supplied request. I'm not thrilled about providing the extra parameter here, but it
	// follows from the net/http/Client.Do(req *Request) from the Go std lib. That is, I don't know the actual request
	// to be performed until runtime and if I am to mock this properly, I don't know the request at the time of
	// injection.
	Execute(req *http.Request)
}

type httpRequester struct {
	client http.Client
	ctx    context.Context
	logger logger.LoggingClient
}

func NewRequester(key RequestType, logger logger.LoggingClient, ctx context.Context) (r Requester, err error) {
	switch key {
	case Http:
		return httpRequester{client: http.Client{}, ctx: ctx, logger: logger}, nil
	case Mock:
		return mockRequester{logger: logger}, nil
	default:
		return nil, fmt.Errorf("unrecognized RequestType value %v", key)
	}
}

func (op httpRequester) Execute(req *http.Request) {
	resp, err := op.client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		return
	}
	op.logger.Error(err.Error())
}

type mockRequester struct {
	logger logger.LoggingClient
}

func (op mockRequester) Execute(req *http.Request) {

}
