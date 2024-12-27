//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"github.com/edgexfoundry/edgex-go/internal/io"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

type AuthController struct {
	dic    *di.Container
	lc     logger.LoggingClient
	reader io.DtoReader
}

func NewAuthController(dic *di.Container) *AuthController {
	lc := container.LoggingClientFrom(dic.Get)

	return &AuthController{
		lc:     lc,
		dic:    dic,
		reader: io.NewJsonDtoReader(),
	}
}
