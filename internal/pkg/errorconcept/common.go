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
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var Common commonErrorConcept

// CommonErrorConcept represents error concepts which apply across core-services
type commonErrorConcept struct {
	ContractInvalid_StatusBadRequest        contractInvalid_StatusBadRequest
	DeleteError                             deleteError
	DuplicateName                           duplicateName
	InvalidRequest_StatusBadRequest         invalidRequest_BadRequest
	InvalidRequest_StatusServiceUnavailable invalidRequest_StatusServiceUnavailable
	ItemNotFound                            itemNotFound
	LimitExceeded                           errLimitExceeded
	RetrieveError_StatusInternalServer      retrieveError_StatusInternalServer
	RetrieveError_StatusServiceUnavailable  retrieveError_ServiceUnavailable
	UpdateError_StatusInternalServer        updateError_StatusInternalServer
	UpdateError_StatusServiceUnavailable    updateError_StatusServiceUnavailable
}

type contractInvalid_StatusBadRequest struct{}

func (r contractInvalid_StatusBadRequest) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r contractInvalid_StatusBadRequest) isA(err error) bool {
	_, ok := err.(models.ErrContractInvalid)
	return ok
}

func (r contractInvalid_StatusBadRequest) message(err error) string {
	return err.Error()
}

type deleteError struct{}

func (r deleteError) httpErrorCode() int {
	return http.StatusServiceUnavailable
}

func (r deleteError) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deleteError) message(err error) string {
	return err.Error()
}

type duplicateName struct{}

func (r duplicateName) httpErrorCode() int {
	return http.StatusConflict
}

func (r duplicateName) isA(err error) bool {
	_, ok := err.(errors.ErrDuplicateName)
	return ok
}

func (r duplicateName) message(err error) string {
	return err.Error()
}

type invalidRequest_BadRequest struct{}

func (r invalidRequest_BadRequest) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r invalidRequest_BadRequest) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r invalidRequest_BadRequest) message(err error) string {
	return err.Error()
}

type invalidRequest_StatusServiceUnavailable struct{}

func (r invalidRequest_StatusServiceUnavailable) httpErrorCode() int {
	return http.StatusServiceUnavailable
}

func (r invalidRequest_StatusServiceUnavailable) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r invalidRequest_StatusServiceUnavailable) message(err error) string {
	return err.Error()
}

type itemNotFound struct{}

func (r itemNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r itemNotFound) isA(err error) bool {
	_, ok := err.(errors.ErrItemNotFound)
	return ok
}

func (r itemNotFound) message(err error) string {
	return err.Error()
}

type errLimitExceeded struct{}

func (r errLimitExceeded) httpErrorCode() int {
	return http.StatusRequestEntityTooLarge
}

func (r errLimitExceeded) isA(err error) bool {
	_, ok := err.(errors.ErrLimitExceeded)
	return ok
}

func (r errLimitExceeded) message(err error) string {
	return err.Error()
}

type retrieveError_StatusInternalServer struct{}

func (r retrieveError_StatusInternalServer) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r retrieveError_StatusInternalServer) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r retrieveError_StatusInternalServer) message(err error) string {
	return err.Error()
}

type retrieveError_ServiceUnavailable struct{}

func (r retrieveError_ServiceUnavailable) httpErrorCode() int {
	return http.StatusServiceUnavailable
}

func (r retrieveError_ServiceUnavailable) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r retrieveError_ServiceUnavailable) message(err error) string {
	return err.Error()
}

type updateError_StatusInternalServer struct{}

func (r updateError_StatusInternalServer) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r updateError_StatusInternalServer) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r updateError_StatusInternalServer) message(err error) string {
	return err.Error()
}

type updateError_StatusServiceUnavailable struct{}

func (r updateError_StatusServiceUnavailable) httpErrorCode() int {
	return http.StatusServiceUnavailable
}

func (r updateError_StatusServiceUnavailable) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r updateError_StatusServiceUnavailable) message(err error) string {
	return err.Error()
}
