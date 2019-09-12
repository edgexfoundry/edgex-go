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

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
)

var Addressable addressableErrorConcept

// AddressableErrorConcept encapsulates error concepts which pertain to addressables
type addressableErrorConcept struct {
	EmptyName                           addressableEmptyName
	InUse                               addressableInUse
	InvalidRequest_StatusInternalServer addressableInvalidRequest_StatusInternalServer
	NotFound                            addressableNotFound
}

type addressableEmptyName struct{}

func (r addressableEmptyName) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r addressableEmptyName) isA(err error) bool {
	_, ok := err.(errors.ErrEmptyAddressableName)
	return ok
}

func (r addressableEmptyName) message(err error) string {
	return err.Error()
}

type addressableInUse struct{}

func (r addressableInUse) httpErrorCode() int {
	return http.StatusConflict
}

func (r addressableInUse) isA(err error) bool {
	_, ok := err.(errors.ErrAddressableInUse)
	return ok
}

func (r addressableInUse) message(err error) string {
	return err.Error()
}

type addressableNotFound struct{}

func (r addressableNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r addressableNotFound) isA(err error) bool {
	_, ok := err.(errors.ErrAddressableNotFound)
	return ok
}

func (r addressableNotFound) message(err error) string {
	return err.Error()
}

type addressableInvalidRequest_StatusInternalServer struct{}

func (r addressableInvalidRequest_StatusInternalServer) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r addressableInvalidRequest_StatusInternalServer) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r addressableInvalidRequest_StatusInternalServer) message(err error) string {
	return err.Error()
}
