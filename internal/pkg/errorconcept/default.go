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
)

var Default defaultErrorConcept

// DefaultErrorConcept represents a fallback error concept only
type defaultErrorConcept struct {
	BadRequest            badRequest
	RequestEntityTooLarge requestEntityTooLarge
	InternalServerError   internalServerError
	ServiceUnavailable    serviceUnavailable
	NotFound              notFound
	Conflict              conflict
}

type badRequest struct{}

func (r badRequest) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r badRequest) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r badRequest) message(err error) string {
	return err.Error()
}

type requestEntityTooLarge struct{}

func (r requestEntityTooLarge) httpErrorCode() int {
	return http.StatusRequestEntityTooLarge
}

func (r requestEntityTooLarge) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r requestEntityTooLarge) message(err error) string {
	return err.Error()
}

type internalServerError struct{}

func (r internalServerError) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r internalServerError) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r internalServerError) message(err error) string {
	return err.Error()
}

type serviceUnavailable struct{}

func (r serviceUnavailable) httpErrorCode() int {
	return http.StatusServiceUnavailable
}

func (r serviceUnavailable) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r serviceUnavailable) message(err error) string {
	return err.Error()
}

type notFound struct{}

func (r notFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r notFound) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r notFound) message(err error) string {
	return err.Error()
}

type conflict struct{}

func (r conflict) httpErrorCode() int {
	return http.StatusConflict
}

func (r conflict) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r conflict) message(err error) string {
	return err.Error()
}
