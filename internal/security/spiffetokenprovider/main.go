/*******************************************************************************
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
 *******************************************************************************/

package spiffetokenprovider

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/spiffetokenprovider/config"
	"github.com/edgexfoundry/edgex-go/internal/security/spiffetokenprovider/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

func Main(ctx context.Context, cancel context.CancelFunc) {
	startupTimer := startup.NewStartUpTimer(common.SecuritySpiffeTokenProviderKey)

	f := flags.New()
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
		common.SecuritySpiffeTokenProviderKey,
		common.ConfigStemSecurity,
		configuration,
		nil,
		startupTimer,
		dic,
		true,
		bootstrapConfig.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			NewBootstrap().BootstrapHandler,
		},
	)

	if !success {
		os.Exit(1)
	}
}
