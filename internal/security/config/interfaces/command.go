//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package interfaces

const (
	// StatusCodeExitNormal exit code
	StatusCodeExitNormal = 0
	// StatusCodeNoOptionSelected exit code
	StatusCodeNoOptionSelected = 1
	// StatusCodeExitWithError is exit code for error
	StatusCodeExitWithError = 2
)

// Command implement the Command pattern
type Command interface {
	Execute() (statusCode int, err error)
}
