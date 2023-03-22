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
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-configuration/v3/configuration"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"gopkg.in/yaml.v3"
)

const (
	commonConfigDone = "IsCommonConfigReady"
)

func Main(ctx context.Context, cancel context.CancelFunc) {
	startupTimer := startup.NewStartUpTimer(common.CoreCommonConfigServiceKey)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be added here,
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

	// need to use in-line function to set the callback type for getAccessToken used in CreateProviderClient to allow
	// access to the config provider in secure mode
	getAccessToken := func() (string, error) {
		accessToken, err := secretProvider.GetAccessToken("consul", common.CoreCommonConfigServiceKey)
		if err != nil {
			return "", fmt.Errorf("failed to get Configuration Provider access token: %s", err.Error())
		}
		lc.Infof("Got Config Provider Access Token with length %d", len(accessToken))
		return accessToken, err
	}

	// create config client
	envVars := environment.NewVariables(lc)
	configProviderInfo, err := config.NewProviderInfo(envVars, f.ConfigProviderUrl())
	if err != nil {
		lc.Errorf("failed to get Provider Info for the common configuration: %s", err.Error())
		os.Exit(1)
	}
	configClient, err := config.CreateProviderClient(lc, common.CoreCommonConfigServiceKey, common.ConfigStemCore, getAccessToken, configProviderInfo.ServiceConfig())
	if err != nil {
		lc.Errorf("failed to create provider client for the common configuration: %s", err.Error())
		os.Exit(1)
	}

	hasConfig := false
	hasConfigSuccess := false
	for startupTimer.HasNotElapsed() {
		// check to see if the configuration exists in the config provider
		hasConfig, err = configClient.HasConfiguration()
		if err == nil {
			hasConfigSuccess = true
			break
		}

		lc.Warnf("Unable to determine if common configuration exists in the provider, will try again: %v", err)
		startupTimer.SleepForInterval()
	}

	if !hasConfigSuccess {
		lc.Errorf("failed to determine if common configuration exists in the provider: %s", err.Error())
		os.Exit(1)
	}

	// load the yaml file and push it using the config client
	if !hasConfig || f.OverwriteConfig() {
		lc.Info("Pushing common configuration. It doesn't exists or overwrite flag is set")

		yamlFile := config.GetConfigFileLocation(lc, f)
		err = pushConfiguration(lc, yamlFile, configClient)
		if err != nil {
			lc.Error(err.Error())
			os.Exit(1)
		}
	} else {
		lc.Info("Skipped pushing common configuration. It already exists and overwrite flag not set")
	}

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

func pushConfiguration(lc logger.LoggingClient, yamlFile string, configClient configuration.Client) error {
	// push not done flag to configClient
	err := configClient.PutConfigurationValue(commonConfigDone, []byte("false"))
	if err != nil {
		return fmt.Errorf("failed to push %s on startup: %s", commonConfigDone, err.Error())
	}
	lc.Infof("Using common configuration from %s", yamlFile)
	contents, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("failed to read common configuration file %s: %s", yamlFile, err.Error())
	}

	data := make(map[string]interface{})
	kv := make(map[string]interface{})

	err = yaml.Unmarshal(contents, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshall common configuration file %s: %s", yamlFile, err.Error())
	}

	kv = buildKeyValues(data, kv, "")

	kv, err = applyEnvOverrides(kv, lc)
	if err != nil {
		return fmt.Errorf("failed to apply env overrides to common configuration: %s", err.Error())
	}

	keys := make([]string, 0, len(kv))

	for k := range kv {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := kv[k]
		// Push key/value into Consul if it is not empty
		if v != nil {
			err = configClient.PutConfigurationValue(k, []byte(fmt.Sprint(v)))
		}
		if err != nil {
			return fmt.Errorf("failed to push common configuration key %s with value %v: %s", k, v, err.Error())
		}
	}

	// push done flag to config client
	err = configClient.PutConfigurationValue(commonConfigDone, []byte("true"))
	if err != nil {
		return fmt.Errorf("failed to push %s on completion: %s", commonConfigDone, err.Error())
	}

	lc.Info("Common configuration has been pushed to into Configuration Provider with overrides applied")

	return nil
}

// buildKeyValues is a helper function to parse the configuration yaml file contents
func buildKeyValues(data map[string]interface{}, kv map[string]interface{}, origKey string) map[string]interface{} {
	key := origKey
	for k, v := range data {
		if len(key) == 0 {
			key = fmt.Sprint(k)
		} else {
			key = fmt.Sprintf("%s/%s", key, k)
		}

		vdata, ok := v.(map[string]interface{})
		if !ok {
			kv[key] = v
			key = origKey
			continue
		}

		kv = buildKeyValues(vdata, kv, key)
		key = origKey
	}

	return kv
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
