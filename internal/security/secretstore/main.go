/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2023 Intel Corporation
 * Copyright 2024-2025 IOTech Ltd
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
 * @author: Alain Pulluelo, ForgeRock AS
 * @author: Tingyu Zeng, Dell
 * @author: Daniel Harms, Dell
 *
 *******************************************************************************/

package secretstore

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	serverConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/server"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/labstack/echo/v4"
)

func Main(ctx context.Context, cancel context.CancelFunc, args []string) {
	startupTimer := startup.NewStartUpTimer(common.SecuritySecretStoreSetupServiceKey)

	var insecureSkipVerify bool
	var secretStoreInterval int
	var longRun bool

	// All common command-line flags have been moved to bootstrap. Service specific flags are add here,
	// but DO NOT call flag.Parse() as it is called by bootstrap.Run() below
	// Service specific used is passed below.
	f := flags.NewWithUsage(
		"    --insecureSkipVerify=true/false Indicates if skipping the server side SSL cert verification, similar to -k of curl\n" +
			"    --secretStoreInterval=<seconds>       Indicates how long the program will pause between the secret store initialization attempts until it succeeds" +
			"    --longRun=true/false                  Indicates whether secret-store-setup is a long run service listening on the server port",
	)

	if len(args) < 1 {
		f.Help()
	}

	f.FlagSet.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "")
	f.FlagSet.IntVar(&secretStoreInterval, "secretStoreInterval", 30, "")
	f.FlagSet.BoolVar(&longRun, "longRun", false, "")
	f.Parse(args)

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	if longRun {
		serverConfig.Configure(ctx, cancel, f, echo.New())
		return
	}

	_, _, success := bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		f,
		common.SecuritySecretStoreSetupServiceKey,
		common.ConfigStemSecurity,
		configuration,
		nil,
		startupTimer,
		dic,
		false,
		bootstrapConfig.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			NewBootstrap(insecureSkipVerify, secretStoreInterval).BootstrapHandler,
		},
	)

	if !success {
		os.Exit(1)
	}
}
