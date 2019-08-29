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

package startup

import (
	"net/http"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/gorilla/mux"
)

// StartHTTPServer creates a new server designed to receive requests from the service defined by the parameters.
func StartHTTPServer(logger logger.LoggingClient, timeout int, router *mux.Router, url string, errChan chan error) {
	go func() {
		correlation.LoggingClient = logger //Not thrilled about this, can't think of anything better ATM
		timeout := time.Millisecond * time.Duration(timeout)

		server := &http.Server{
			Handler:      router,
			Addr:         url,
			WriteTimeout: timeout,
			ReadTimeout:  timeout,
		}

		errChan <- server.ListenAndServe()
	}()
}
