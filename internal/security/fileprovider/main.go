/*******************************************************************************
 * Copyright 2023 Intel Corporation
 * Copyright (C) 2025 IOTech Ltd
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
 *
 *******************************************************************************/

package fileprovider

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/command"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/command/handlers"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
)

func Main(ctx context.Context, cancel context.CancelFunc, args []string) {
	startupTimer := startup.NewStartUpTimer(common.SecurityFileTokenProviderServiceKey)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be added here,
	// by inserting service specific flag prior to call to commonFlags.Parse().
	// Example:
	// 		flags.FlagSet.StringVar(&myvar, "m", "", "Specify a ....")
	//      ....
	//      flags.Parse(args)
	//
	f := command.NewCommonFlags()
	f.Parse(args)

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	bootStrapper := NewBootstrap()
	cmdBootStrapper := handlers.NewInitialization()

	_, _, success := bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		f,
		common.SecurityFileTokenProviderServiceKey,
		common.ConfigStemSecurity,
		configuration,
		nil,
		startupTimer,
		dic,
		false,
		bootstrapConfig.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			bootStrapper.BootstrapHandler,
			cmdBootStrapper.BootstrapHandler,
		},
	)

	if !success {
		os.Exit(1)
	}
}
