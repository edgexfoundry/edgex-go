//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/error"

	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

// ErrorHandler contains the name of the error.Handler implementation in the DIC.
var ErrorHandlerName = di.TypeInstanceToName(error.Handler{})

// ErrorHandlerFrom helper function queries the DIC and returns the error.Handler implementation.
func ErrorHandlerFrom(get di.Get) *error.Handler {
	return get(ErrorHandlerName).(*error.Handler)
}
