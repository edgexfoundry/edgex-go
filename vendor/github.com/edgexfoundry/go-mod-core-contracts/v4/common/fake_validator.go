//go:build no_dto_validator

//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import "errors"

func Validate(a interface{}) error {
	return errors.New("wrong build: DTO validator is not available. " +
		"Build without \"-tags no_dto_validator\" on the go build command line to enable runtime support for this feature")
}
