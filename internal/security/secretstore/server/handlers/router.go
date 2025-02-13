//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/controller"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/labstack/echo/v4"
)

// TODO: this will be declared in go-mod-core-contracts after the review
const regenTokenRoute = "/api/v3/token/entityId/:entityId"

// LoadRestRoutes generates the routing for API requests
// Authentication is always on for this service,
// as it is called by NGINX to authenticate requests
// and must always authenticate even if the rest of EdgeX does not
func LoadRestRoutes(r *echo.Echo, dic *di.Container) {
	ac := controller.NewTokenController(dic)
	r.PUT(regenTokenRoute, ac.RegenToken)
}
