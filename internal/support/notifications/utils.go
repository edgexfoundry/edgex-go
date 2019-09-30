/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package notifications

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

const (
	ExceededMaxResultCount string = "error, exceeded the max limit as defined in config"
)

// Printing function purely for debugging purposes
// Print the body of a request to the console
func printBody(r io.ReadCloser) {
	body, err := ioutil.ReadAll(r)
	bodyString := string(body)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(bodyString)
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(clients.ContentType, clients.ContentTypeText)
	w.Write([]byte("pong"))
}

func checkMaxLimit(limit int) error {
	if limit > Configuration.Service.MaxResultCount {
		LoggingClient.Error(ExceededMaxResultCount)
		return errors.NewErrLimitExceeded(limit)
	}

	return nil
}
