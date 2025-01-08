//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"github.com/edgexfoundry/edgex-go/internal/io"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

type AuthController struct {
	dic    *di.Container
	reader io.DtoReader
}

func NewAuthController(dic *di.Container) *AuthController {
	return &AuthController{
		dic:    dic,
		reader: io.NewJsonDtoReader(),
	}
}
