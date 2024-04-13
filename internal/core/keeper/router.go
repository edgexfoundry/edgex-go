//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	keeperController "github.com/edgexfoundry/edgex-go/internal/core/keeper/controller/http"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// KV
	kv := keeperController.NewKVController(dic)
	r.GET(common.ApiKVSByKeyEchoRoute, kv.Keys, authenticationHook)
	r.PUT(common.ApiKVSByKeyEchoRoute, kv.AddKeys, authenticationHook)
	r.DELETE(common.ApiKVSByKeyEchoRoute, kv.DeleteKeys, authenticationHook)

	// Registry
	rc := keeperController.NewRegistryController(dic)
	r.POST(common.ApiRegisterRoute, rc.Register, authenticationHook)
	r.PUT(common.ApiRegisterRoute, rc.UpdateRegister, authenticationHook)
	r.GET(common.ApiAllRegistrationsRoute, rc.Registrations, authenticationHook)
	r.GET(common.ApiRegistrationByServiceIdEchoRoute, rc.RegistrationByServiceId, authenticationHook)
	r.DELETE(common.ApiRegistrationByServiceIdEchoRoute, rc.Deregister, authenticationHook)
}
