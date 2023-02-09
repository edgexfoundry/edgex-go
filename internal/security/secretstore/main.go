/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
)

func Main(ctx context.Context, cancel context.CancelFunc) {
	startupTimer := startup.NewStartUpTimer(common.SecuritySecretStoreSetupServiceKey)

	var insecureSkipVerify bool
	var vaultInterval int

	// All common command-line flags have been moved to bootstrap. Service specific flags are add here,
	// but DO NOT call flag.Parse() as it is called by bootstrap.Run() below
	// Service specific used is passed below.
	f := flags.NewWithUsage(
		"    --insecureSkipVerify=true/false Indicates if skipping the server side SSL cert verification, similar to -k of curl\n" +
			"    --vaultInterval=<seconds>       Indicates how long the program will pause between vault initialization attempts until it succeeds",
	)

	if len(os.Args) < 2 {
		f.Help()
	}

	f.FlagSet.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "")
	f.FlagSet.IntVar(&vaultInterval, "vaultInterval", 30, "")
	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

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
			NewBootstrap(insecureSkipVerify, vaultInterval).BootstrapHandler,
		},
	)

	if !success {
		os.Exit(1)
	}
}
