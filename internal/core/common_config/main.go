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
 *******************************************************************************/

package common_config

import (
	"context"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func Main(ctx context.Context, cancel context.CancelFunc) {

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be add here,
	// by inserting service specific flag prior to call to commonFlags.Parse().
	// Example:
	// 		flags.FlagSet.StringVar(&myvar, "m", "", "Specify a ....")
	//      ....
	//      flags.Parse(os.Args[1:])
	//

	// TODO: figure out how to eliminate registry and profile flags
	f := flags.New()
	f.Parse(os.Args[1:])

	var wg sync.WaitGroup
	translateInterruptToCancel(ctx, &wg, cancel)

	lc := logger.NewClient(common.CoreCommonConfigServiceKey, models.InfoLog)
	lc.Info("Core Common Config is starting")
	startupTimer := startup.NewStartUpTimer(common.CoreCommonConfigServiceKey)
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	secretProvider, err := secret.NewSecretProvider(nil, environment.NewVariables(lc), ctx, startupTimer, dic, common.CoreCommonConfigServiceKey)
	if err != nil {
		lc.Errorf("failed to create Secret Provider: %v", err)
		os.Exit(1)
	}

	lc.Info("Secret Provider created")

	_, err = secretProvider.GetAccessToken("consul", common.CoreCommonConfigServiceKey)
	if err != nil {
		lc.Errorf("failed to get Access Token for config provider: %v", err)
		os.Exit(1)
	}

	lc.Info("Got Config Provider Access Token")
	lc.Info("Core Common Config Ready for stage two")
	lc.Info("Core Common Config exiting")
	os.Exit(0)
}

// translateInterruptToCancel spawns a go routine to translate the receipt of a SIGTERM signal to a call to cancel
// the context used by the bootstrap implementation.
func translateInterruptToCancel(ctx context.Context, wg *sync.WaitGroup, cancel context.CancelFunc) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		signalStream := make(chan os.Signal, 1)
		defer func() {
			signal.Stop(signalStream)
			close(signalStream)
		}()
		signal.Notify(signalStream, os.Interrupt, syscall.SIGTERM)
		select {
		case <-signalStream:
			cancel()
			return
		case <-ctx.Done():
			return
		}
	}()
}
