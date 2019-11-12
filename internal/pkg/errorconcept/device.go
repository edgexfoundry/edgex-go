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

	command "github.com/edgexfoundry/edgex-go/internal/core/command/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

var Device deviceErrorConcept

// DeviceErrorConcept represents the accessor for the device-specific error concepts
type deviceErrorConcept struct {
	Locked         deviceLocked
	NotFound       deviceNotFound
	NotFoundInDB   deviceNotFoundInDB
	NotifyError    deviceNotify
	RequesterError deviceRequester
}

type deviceLocked struct{}

func (r deviceLocked) httpErrorCode() int {
	return http.StatusNotFound
}

func (r deviceLocked) isA(err error) bool {
	_, ok := err.(command.ErrDeviceLocked)
	return ok
}

func (r deviceLocked) message(err error) string {
	return err.Error()
}



type deviceNotFound struct{}

func (r deviceNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r deviceNotFound) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceNotFound) message(err error) string {
	return err.Error()
}

type deviceNotFoundInDB struct{}

func (r deviceNotFoundInDB) httpErrorCode() int {
	return http.StatusNotFound
}

func (r deviceNotFoundInDB) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r deviceNotFoundInDB) message(err error) string {
	return "Device not found"
}

type deviceNotify struct{}

func (r deviceNotify) httpErrorCode() int {
	return http.StatusServiceUnavailable
}

func (r deviceNotify) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceNotify) message(err error) string {
	return err.Error()
}

type deviceRequester struct{}

func (r deviceRequester) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r deviceRequester) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceRequester) message(err error) string {
	return err.Error()
}
