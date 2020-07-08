//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package error

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/v2"
)

// NewServiceClientHttpError represents the accessor for the service-client-specific error concepts
func NewServiceClientHttpError(err error) *serviceClientHttpError {
	return &serviceClientHttpError{Err: err}
}

type serviceClientHttpError struct {
	Err error
}

func (r serviceClientHttpError) httpErrorCode() int {
	return r.Err.(types.ErrServiceClient).StatusCode
}

func (r serviceClientHttpError) isA(err error) bool {
	_, ok := err.(types.ErrServiceClient)
	return ok
}

func (r serviceClientHttpError) message(err error) string {
	return err.Error()
}

// NewErrContractInvalid returns an instance of the error interface with ErrContractInvalid as its implementation.
func NewErrContractInvalidError(err error) *errContractInvalid {
	return &errContractInvalid{Err: err}
}

// ErrContractInvalid is a specific error type for handling model validation failures. Type checking within
// the calling application will facilitate more explicit error handling whereby it's clear that validation
// has failed as opposed to something unexpected happening.
type errContractInvalid struct {
	Err error
}

func (r errContractInvalid) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r errContractInvalid) isA(err error) bool {
	_, ok := err.(v2.ErrContractInvalid)
	return ok
}

func (r errContractInvalid) message(err error) string {
	return err.Error()
}
