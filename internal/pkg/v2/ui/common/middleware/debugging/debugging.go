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

// debugging implements middleware that logs all requests and responses.
package debugging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// middleware contains references to dependencies required by the debugging middleware implementation.
type middleware struct {
	lc logger.LoggingClient
}

// New is a factory function that returns a new middleware receiver.
func New(lc logger.LoggingClient) *middleware {
	return &middleware{
		lc: lc,
	}
}

// marshal converts interface{}-typed content into a string.
func (m *middleware) marshal(content interface{}) string {
	s, err := json.Marshal(content)
	if err != nil {
		s = []byte("(unable to marshal)")
	}
	return string(s)
}

// Handler implements the middleware.Handler contract; it logs each request and response.
func (m *middleware) Handler(
	request interface{},
	behavior *application.Behavior,
	execute application.Executable) (response interface{}, status infrastructure.Status) {

	start := time.Now()
	response, status = execute(request)
	elapsed := time.Now().Sub(start) / time.Millisecond

	m.lc.Debug(
		fmt.Sprintf("elapsed: %dms, version: %s, kind: %s, action: %s, request: %s, response: %s",
			elapsed,
			behavior.Version,
			behavior.Kind,
			behavior.Action,
			m.marshal(request),
			m.marshal(response),
		),
	)
	return
}
