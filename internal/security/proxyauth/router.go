//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package proxyauth

import (
	"github.com/edgexfoundry/edgex-go"
	spaController "github.com/edgexfoundry/edgex-go/internal/security/proxyauth/controller"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/labstack/echo/v4"
)

// LoadRestRoutes generates the routing for API requests
// Authentication is always on for this service,
// as it is called by NGINX to authenticate requests
// and must always authenticate even if the rest of EdgeX does not
func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	authenticationHook := handlers.AutoConfigAuthenticationFunc(dic)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Run authentication hook for a nil route
	r.GET("/auth", emptyHandler, authenticationHook)

	ac := spaController.NewAuthController(dic)

	r.POST(common.ApiKeyRoute, ac.AddKey, authenticationHook)
	// This API will be called within the authenticationHook function itself
	r.GET(common.ApiVerificationKeyByIssuerRoute, ac.VerificationKeyByIssuer)
}
