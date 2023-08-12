/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright (c) 2019-2023 Intel Corporation
 * Copyright (C) 2021-2023 IOTech Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package proxyauth

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/labstack/echo/v4"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router      *echo.Echo
	serviceName string
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *echo.Echo, serviceName string) *Bootstrap {
	return &Bootstrap{
		router:      router,
		serviceName: serviceName,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the command service.
// Authentication is always on for this service,
// as it is called by NGINX to authenticate requests
// and must always authenticate even if the rest of EdgeX does not
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.VaultAuthenticationHandlerFunc(secretProvider, lc)

	// Common
	_ = controller.NewCommonController(dic, b.router, b.serviceName, edgex.Version)

	// Run authentication hook for a nil route
	b.router.GET("/auth", emptyHandler, authenticationHook)

	return true
}

func emptyHandler(_ echo.Context) error { return nil }
