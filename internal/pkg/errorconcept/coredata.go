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

package errorconcept

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
)

var CoreData coreDataErrorConcept

// coreDataErrorConcept contains the superfluous, duplicated errors that must be handled for CoreData
type coreDataErrorConcept struct {
	DBNotFound   coreDataDBNotFound
	JsonDecoding coreDataJSONDecoding
}

type coreDataDBNotFound struct{}

func (r coreDataDBNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r coreDataDBNotFound) isA(err error) bool {
	_, ok := err.(errors.ErrDbNotFound)
	return ok
}

func (r coreDataDBNotFound) message(err error) string {
	return err.Error()
}

type coreDataJSONDecoding struct{}

func (r coreDataJSONDecoding) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r coreDataJSONDecoding) isA(err error) bool {
	_, ok := err.(errors.ErrJsonDecoding)
	return ok
}

func (r coreDataJSONDecoding) message(err error) string {
	return err.Error()
}
