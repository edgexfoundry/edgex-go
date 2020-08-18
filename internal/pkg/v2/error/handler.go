//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package error

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type ErrorType interface {
	httpErrorCode() int
	isA(err error) bool
	message(err error) string
}

type Handler struct {
	logger logger.LoggingClient
}

type HTTPErrorResponse struct {
	ErrorCode  uint16
	ErrMessage string
}

type ErrorHandler interface {
	Handle(err error, ec ErrorType) HTTPErrorResponse
	HandleWithDefault(err error, allowableError ErrorType, defaultError ErrorType) HTTPErrorResponse
	HandleManyWithDefault(err error, allowableErrors []ErrorType, defaultError ErrorType) HTTPErrorResponse
}

func NewErrorHandler(l logger.LoggingClient) ErrorHandler {
	h := Handler{l}
	return &h
}

// Handle applies the specified error to the HTTP response writer
func (e *Handler) Handle(err error, ec ErrorType) HTTPErrorResponse {
	message := ec.message(err)
	e.logger.Error(message)
	return HTTPErrorResponse{
		ErrorCode:  uint16(ec.httpErrorCode()),
		ErrMessage: message,
	}

}

// HandleWithDefault applies general error-handling with a single allowable error and a default error to be used as a
// fallback when none of the allowable errors are matched
func (e *Handler) HandleWithDefault(err error, allowableError ErrorType, defaultError ErrorType) HTTPErrorResponse {
	if allowableError != nil && allowableError.isA(err) {
		return e.Handle(err, allowableError)
	}
	return e.Handle(err, defaultError)
}

// HandleManyWithDefault applies general error-handling for the specified set of allowable errors and a default error to be used
// as a fallback when none of the allowable errors are matched
func (e *Handler) HandleManyWithDefault(err error, allowableErrors []ErrorType, defaultError ErrorType) HTTPErrorResponse {
	for key := range allowableErrors {
		if allowableErrors[key].isA(err) {
			return e.Handle(err, allowableErrors[key])
		}
	}
	return e.Handle(err, defaultError)
}
