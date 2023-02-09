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

package bootstrapper

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	bootstrapper "github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/container"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/handlers"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/redis"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

const (
	configureDatabaseSubcommandName = "configureRedis"
	setupMsgbusCredsSubcommandName  = "setupMessageBusCreds"
)

// Main function is the wrapper for the security bootstrapper main
func Main(ctx context.Context, cancel context.CancelFunc) {
	// service key for this bootstrapper service
	startupTimer := startup.NewStartUpTimer(common.SecurityBootstrapperKey)

	// Common Command-line flags have been moved to command.CommonFlags, but this service doesn't use all
	// the common flags so we are using our own implementation of the CommonFlags interface
	f := bootstrapper.NewCommonFlags()

	f.Parse(os.Args[1:])

	// find out the subcommand name before assigning the real concrete configuration
	// bootstrapRedis has its own configuration settings
	var configDir string
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagSet.StringVar(&configDir, "configDir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// branch out to different entrypoints if it is from sub commands
	switch flagSet.Arg(0) {

	case configureDatabaseSubcommandName:
		redis.Configure(ctx, cancel, f)
		return
	case setupMsgbusCredsSubcommandName:
		err = ConfigureSecureMessageBus(flagSet.Arg(1), ctx, cancel, f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	serviceHandler := handlers.NewInitialization()
	bootstrap.Run(
		ctx,
		cancel,
		f,
		common.SecurityBootstrapperKey,
		common.ConfigStemSecurity,
		configuration,
		startupTimer,
		dic,
		false,
		bootstrapConfig.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			serviceHandler.BootstrapHandler,
		},
	)

	// exit with the code specified by serviceHandler
	os.Exit(serviceHandler.GetExitStatusCode())
}
