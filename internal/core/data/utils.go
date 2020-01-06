/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package data

import (
	"io"
	"io/ioutil"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// Printing function purely for debugging purposes
// Print the body of a request to the console
func printBody(r io.ReadCloser, lc logger.LoggingClient) {
	body, err := ioutil.ReadAll(r)
	bodyString := string(body)

	if err != nil {
		lc.Error(err.Error())
	}

	lc.Info(bodyString)
}

func checkMaxLimit(limit int, lc logger.LoggingClient, configuration *config.ConfigurationStruct) error {
	if limit > configuration.Service.MaxResultCount {
		lc.Error(maxExceededString)
		return errors.NewErrLimitExceeded(limit)
	}

	return nil
}
