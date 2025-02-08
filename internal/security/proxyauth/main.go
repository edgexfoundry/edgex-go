/*******************************************************************************
 * Copyright 2020 Dell Inc.
 * Copyright 2022-2025 IOTech Ltd.
 * Copyright 2023 Intel Corporation
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

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/edgex-go"
	pkgHandlers "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/config"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/container"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/embed"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/labstack/echo/v4"
)

func Main(ctx context.Context, cancel context.CancelFunc, router *echo.Echo, args []string) {
	startupTimer := startup.NewStartUpTimer(common.SecurityProxyAuthServiceKey)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be added here,
	// by inserting service specific flag prior to call to commonFlags.Parse().
	// Example:
	// 		flags.FlagSet.StringVar(&myvar, "m", "", "Specify a ....")
	//      ....
	//      flags.Parse(args)
	//
	f := flags.New()
	f.Parse(args)

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		container.CryptoInterfaceName: func(get di.Get) interface{} {
			return crypto.NewAESCryptor()
		},
	})

	httpServer := handlers.NewHttpServer(router, true, common.SecurityProxyAuthServiceKey)
	dbHandler := pkgHandlers.NewDatabase(httpServer, configuration, container.DBClientInterfaceName, embed.SchemaName,
		common.SecurityProxyAuthServiceKey, edgex.Version, embed.SQLFiles)

	bootstrap.Run(
		ctx,
		cancel,
		f,
		common.SecurityProxyAuthServiceKey,
		common.ConfigStemSecurity,
		configuration,
		startupTimer,
		dic,
		true,
		bootstrapConfig.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			handlers.NewClientsBootstrap().BootstrapHandler,
			dbHandler.BootstrapHandler, // add db client bootstrap handler
			NewBootstrap(router, common.SecurityProxyAuthServiceKey).BootstrapHandler,
			httpServer.BootstrapHandler,
			handlers.NewStartMessage(common.SecurityProxyAuthServiceKey, edgex.Version).BootstrapHandler,
		})
}
