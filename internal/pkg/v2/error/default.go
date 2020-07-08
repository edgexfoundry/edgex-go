//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package error

import (
	"net/http"
)

const (
	METHOD_NOT_ALLOWED = "isA should not be invoked, this is to only be used as a default error"
)

var Default defaultError

// defaultError represents a fallback error only
type defaultError struct {
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
