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

var ValueDescriptors valueDescriptorsErrorConcept

// ValueDescriptorsErrorConcept represents the accessor for the value-descriptor-specific error concepts
type valueDescriptorsErrorConcept struct {
	DuplicateName valueDescriptorDuplicateName
	InUse         valueDescriptorsInUse
	Invalid       valueDescriptorInvalid
	LimitExceeded valueDescriptorLimitExceeded
	NotFound      valueDescriptorNotFound
	NotFoundInDB  valueDescriptorDBNotFound
}

type valueDescriptorDuplicateName struct{}

func (r valueDescriptorDuplicateName) httpErrorCode() int {
	return http.StatusConflict
}

func (r valueDescriptorDuplicateName) isA(err error) bool {
	_, ok := err.(errors.ErrDuplicateValueDescriptorName)
	return ok
}

func (r valueDescriptorDuplicateName) message(err error) string {
	return err.Error()
}

type valueDescriptorsInUse struct{}

func (r valueDescriptorsInUse) httpErrorCode() int {
	return http.StatusConflict
}

func (r valueDescriptorsInUse) isA(err error) bool {
	_, ok := err.(errors.ErrValueDescriptorsInUse)
	return ok
}

func (r valueDescriptorsInUse) message(err error) string {
	return err.Error()
}

type valueDescriptorInvalid struct{}

func (r valueDescriptorInvalid) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r valueDescriptorInvalid) isA(err error) bool {
	_, ok := err.(errors.ErrValueDescriptorInvalid)
	return ok
}

func (r valueDescriptorInvalid) message(err error) string {
	return err.Error()
}

type valueDescriptorLimitExceeded struct{}

func (r valueDescriptorLimitExceeded) httpErrorCode() int {
	return http.StatusRequestEntityTooLarge
}

func (r valueDescriptorLimitExceeded) isA(err error) bool {
	_, ok := err.(errors.ErrLimitExceeded)
	return ok
}

func (r valueDescriptorLimitExceeded) message(err error) string {
	return err.Error()
}

type valueDescriptorNotFound struct{}

func (r valueDescriptorNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r valueDescriptorNotFound) isA(err error) bool {
	_, ok := err.(errors.ErrValueDescriptorNotFound)
	return ok
}

func (r valueDescriptorNotFound) message(err error) string {
	return err.Error()
}

type valueDescriptorDBNotFound struct{}

func (r valueDescriptorDBNotFound) httpErrorCode() int {
	return http.StatusConflict
}

func (r valueDescriptorDBNotFound) isA(err error) bool {
	_, ok := err.(errors.ErrDbNotFound)
	return ok
}

func (r valueDescriptorDBNotFound) message(err error) string {
	return "Value descriptor not found for reading"
}
