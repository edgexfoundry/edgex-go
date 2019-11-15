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

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type ErrorConceptType interface {
	httpErrorCode() int
	isA(err error) bool
	message(err error) string
}

type Handler struct {
	logger logger.LoggingClient
}

type ErrorHandler interface {
	Handle(w http.ResponseWriter, err error, ec ErrorConceptType)
	HandleManyVariants(w http.ResponseWriter, err error, allowableErrors []ErrorConceptType, defaultError ErrorConceptType)
	HandleOneVariant(w http.ResponseWriter, err error, allowableError ErrorConceptType, defaultError ErrorConceptType)
}

func NewErrorHandler(l logger.LoggingClient) ErrorHandler {
	h := Handler{l}
	return &h
}

// Handle applies the specified error and error concept tot he HTTP response writer
func (e *Handler) Handle(w http.ResponseWriter, err error, ec ErrorConceptType) {
	message := ec.message(err)
	e.logger.Error(message)
	http.Error(w, message, ec.httpErrorCode())
}

// HandleOneVariant applies general error-handling with a single allowable error and a default error to be used as a
// fallback when none of the allowable errors are matched
func (e *Handler) HandleOneVariant(w http.ResponseWriter, err error, allowableError ErrorConceptType, defaultError ErrorConceptType) {
	if allowableError != nil && allowableError.isA(err) {
		e.Handle(w, err, allowableError)
		return
	}
	e.Handle(w, err, defaultError)
}

// HandleManyVariants applies general error-handling for the specified set of allowable errors and a default error to be used
// as a fallback when none of the allowable errors are matched
func (e *Handler) HandleManyVariants(w http.ResponseWriter, err error, allowableErrors []ErrorConceptType, defaultError ErrorConceptType) {
	for key := range allowableErrors {
		if allowableErrors[key].isA(err) {
			e.Handle(w, err, allowableErrors[key])
			return
		}
	}
	e.Handle(w, err, defaultError)
}
