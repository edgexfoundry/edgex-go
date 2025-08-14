/*******************************************************************************
 * Copyright 2023 Intel Corporation
 * Copyright 2025 IOTech Ltd
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
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/file"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

const (
	commonConfigDone = "IsCommonConfigReady"
)

func Main(ctx context.Context, cancel context.CancelFunc, args []string) {
	startupTimer := startup.NewStartUpTimer(common.CoreCommonConfigServiceKey)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be added here,
	// by inserting service specific flag prior to call to commonFlags.Parse().
	// Example:
	// 		flags.FlagSet.StringVar(&myvar, "m", "", "Specify a ....")
	//      ....
	//      flags.Parse(args)
	//

	// TODO: figure out how to eliminate registry and profile flags
	f := flags.New()
	f.Parse(args)

	var wg sync.WaitGroup
	translateInterruptToCancel(ctx, &wg, cancel)

	lc := logger.NewClient(common.CoreCommonConfigServiceKey, models.InfoLog)
	lc.Info("Core Common Config is starting")
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

	// create config client
	envVars := environment.NewVariables(lc)
	configProviderInfo, err := config.NewProviderInfo(envVars, f.ConfigProviderUrl())
	if err != nil {
		lc.Errorf("failed to get Provider Info for the common configuration: %s", err.Error())
		os.Exit(1)
	}

	// Bypass the zero trust zitidfied transport for Core Keeper Configuration client
	// Should leverage the HttpTransportFromService function from zerotrust pkg in go-mod-bootstrap, set the default transport for now
	secretProvider.SetHttpTransport(http.DefaultTransport)
	jwtSecretProvider := secret.NewJWTSecretProvider(secretProvider)

	configProviderInfo.SetAuthInjector(jwtSecretProvider)

	configClient, err := config.CreateProviderClient(lc, common.CoreCommonConfigServiceKey, common.ConfigStemCore, configProviderInfo.ServiceConfig())
	if err != nil {
		lc.Errorf("failed to create provider client for the common configuration: %s", err.Error())
		os.Exit(1)
	}

	// push not done flag to configClient
	err = configClient.PutConfigurationValue(commonConfigDone, []byte(common.ValueFalse))
	if err != nil {
		lc.Errorf("failed to push %s on startup: %s", commonConfigDone, err.Error())
		os.Exit(1)
	}

	yamlFile := config.GetConfigFileLocation(lc, f)
	lc.Infof("Using common configuration from %s", yamlFile)

	configMap, err := loadConfigMapFromYamlFile(yamlFile, secretProvider, lc)
	if err != nil {
		lc.Errorf("Failed to load %s : %s", yamlFile, err.Error())
		os.Exit(1)
	}
	configMap, err = applyEnvOverrides(configMap, lc)
	if err != nil {
		lc.Errorf("Failed to apply env overrides to common configuration: %s", err.Error())
		os.Exit(1)
	}

	// PutConfigurationMap func will check each keys and put the config value if not exist or override is set
	err = configClient.PutConfigurationMap(configMap, getOverwriteConfig(f, lc))
	if err != nil {
		lc.Errorf("Failed to put configuration: %s", err.Error())
		os.Exit(1)
	}

	// push done flag to config client
	err = configClient.PutConfigurationValue(commonConfigDone, []byte(common.ValueTrue))
	if err != nil {
		lc.Errorf("Failed to push %s on completion: %s", commonConfigDone, err.Error())
		os.Exit(1)
	}

	lc.Info("Core Common Config exiting")
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

func applyEnvOverrides(keyValues map[string]any, lc logger.LoggingClient) (map[string]any, error) {
	env := environment.NewVariables(lc)

	overrideCount, err := env.OverrideConfigMapValues(keyValues)
	if err != nil {
		return nil, err
	}

	lc.Infof("Common configuration loaded from file with %d overrides applied", overrideCount)

	return keyValues, nil
}

func loadConfigMapFromYamlFile(yamlFile string, secretProvider interfaces.SecretProvider, lc logger.LoggingClient) (map[string]any, error) {
	contents, err := file.Load(yamlFile, secretProvider, lc)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %s", yamlFile, err.Error())
	}

	data := make(map[string]any)

	err = yaml.Unmarshal(contents, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall yaml configuration: %s", err.Error())
	}
	return data, nil
}
