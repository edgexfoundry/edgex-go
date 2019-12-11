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
	"io"
	"io/ioutil"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	notificationsConfig "github.com/edgexfoundry/edgex-go/internal/support/notifications/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
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

func checkMaxLimit(limit int, loggingClient logger.LoggingClient, config notificationsConfig.ConfigurationStruct) error {
	if limit > config.Service.MaxResultCount {
		loggingClient.Error(ExceededMaxResultCount)
		return errors.NewErrLimitExceeded(limit)
	}

	return nil
}
