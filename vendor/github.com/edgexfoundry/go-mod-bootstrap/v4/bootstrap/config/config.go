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
 *******************************************************************************/

package config

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/file"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/mitchellh/copystructure"
	"gopkg.in/yaml.v3"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-configuration/v4/configuration"
	"github.com/edgexfoundry/go-mod-configuration/v4/pkg/types"

	clientinterfaces "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
)

const (
	writableKey        = "Writable"
	insecureSecretsKey = "InsecureSecrets"
	secretNameKey      = "SecretName"
	secretDataKey      = "SecretData"

	allServicesKey    = "all-services"
	appServicesKey    = "app-services"
	deviceServicesKey = "device-services"

	SecurityModeKey = "Mode"
)

var invalidRemoteHostsError = errors.New("-rsh/--remoteServiceHosts must contain 3 and only 3 comma seperated host names")

// UpdatedStream defines the stream type that is notified by ListenForChanges when a configuration update is received.
type UpdatedStream chan struct{}

type Processor struct {
	lc                  logger.LoggingClient
	flags               flags.Common
	envVars             *environment.Variables
	startupTimer        startup.Timer
	ctx                 context.Context
	wg                  *sync.WaitGroup
	configUpdated       UpdatedStream
	dic                 *di.Container
	overwriteConfig     *bool
	providerHasConfig   bool
	commonConfigClient  configuration.Client
	appConfigClient     configuration.Client
	deviceConfigClient  configuration.Client
	privateConfigClient configuration.Client
}

// NewProcessor creates a new configuration Processor
func NewProcessor(
	flags flags.Common,
	envVars *environment.Variables,
	startupTimer startup.Timer,
	ctx context.Context,
	wg *sync.WaitGroup,
	configUpdated UpdatedStream,
	dic *di.Container,
) *Processor {
	return &Processor{
		lc:            container.LoggingClientFrom(dic.Get),
		flags:         flags,
		envVars:       envVars,
		startupTimer:  startupTimer,
		ctx:           ctx,
		wg:            wg,
		configUpdated: configUpdated,
		dic:           dic,
	}
}

func NewProcessorForCustomConfig(
	flags flags.Common,
	ctx context.Context,
	wg *sync.WaitGroup,
	dic *di.Container) *Processor {

	privateConfigClient := container.ConfigClientFrom(dic.Get)

	return &Processor{
		lc:                  container.LoggingClientFrom(dic.Get),
		flags:               flags,
		ctx:                 ctx,
		wg:                  wg,
		dic:                 dic,
		privateConfigClient: privateConfigClient,
	}
}

func (cp *Processor) Process(
	serviceKey string,
	serviceType string,
	configStem string,
	serviceConfig interfaces.Configuration,
	secretProvider interfaces.SecretProviderExt,
	jwtSecretProvider clientinterfaces.AuthenticationInjector) error {

	configProviderUrl := cp.flags.ConfigProviderUrl()
	remoteHosts := environment.GetRemoteServiceHosts(cp.lc, cp.flags.RemoteServiceHosts())

	// Create new ProviderInfo and initialize it from command-line flag or Variables
	configProviderInfo, err := NewProviderInfo(cp.envVars, configProviderUrl)
	if err != nil {
		return err
	}

	configProviderInfo.SetAuthInjector(jwtSecretProvider)

	useProvider := configProviderInfo.UseProvider()

	mode := &container.DevRemoteMode{
		InDevMode:    cp.flags.InDevMode(),
		InRemoteMode: remoteHosts != nil,
	}

	cp.dic.Update(di.ServiceConstructorMap{
		container.DevRemoteModeName: func(get di.Get) interface{} {
			return mode
		},
	})

	if useProvider {
		if remoteHosts != nil {
			if len(remoteHosts) != 3 {
				return invalidRemoteHostsError
			}

			cp.lc.Infof("Setting config Provider host to %s", remoteHosts[1])
			configProviderInfo.SetHost(remoteHosts[1])
		}

		if err := cp.loadConfigByProvider(
			configProviderInfo,
			configStem,
			serviceConfig,
			serviceType,
			serviceKey); err != nil {
			return err
		}

		// listen for changes on Writable
		cp.listenForPrivateChanges(serviceConfig, cp.privateConfigClient, utils.BuildBaseKey(configStem, serviceKey))
		cp.lc.Infof("listening for private config changes")
		cp.listenForCommonChanges(serviceConfig, cp.commonConfigClient, cp.privateConfigClient, utils.BuildBaseKey(configStem, common.CoreCommonConfigServiceKey, allServicesKey))
		cp.lc.Infof("listening for all services common config changes")
		if cp.appConfigClient != nil {
			cp.listenForCommonChanges(serviceConfig, cp.appConfigClient, cp.privateConfigClient, utils.BuildBaseKey(configStem, common.CoreCommonConfigServiceKey, appServicesKey))
			cp.lc.Infof("listening for application service common config changes")
		}
		if cp.deviceConfigClient != nil {
			cp.listenForCommonChanges(serviceConfig, cp.deviceConfigClient, cp.privateConfigClient, utils.BuildBaseKey(configStem, common.CoreCommonConfigServiceKey, deviceServicesKey))
			cp.lc.Infof("listening for device service common config changes")
		}
	} else {
		// Now load common configuration from local file if not using config provider and -cc/--commonConfig flag is used.
		// NOTE: Some security services don't use any common configuration and don't use the configuration provider.
		commonConfigLocation := environment.GetCommonConfigFileName(cp.lc, cp.flags.CommonConfig())
		if commonConfigLocation != "" {
			err := cp.loadCommonConfigFromFile(commonConfigLocation, serviceConfig, serviceType)
			if err != nil {
				return err
			}

			overrideCount, err := cp.envVars.OverrideConfiguration(serviceConfig)
			if err != nil {
				return err
			}
			cp.lc.Infof("Common configuration loaded from file with %d overrides applied", overrideCount)
		}

		if _, err := cp.loadPrivateConfigFromFile(serviceConfig); err != nil {
			return err
		}

	}

	// Now that configuration has been loaded and overrides applied the log level can be set as configured.
	err = cp.lc.SetLogLevel(serviceConfig.GetLogLevel())

	if cp.flags.InDevMode() {
		// Dev mode is for when running service with Config Provider in hybrid mode (all other service running in Docker).
		// All the host values are set to the docker names in the common configuration, so must be overridden here with "localhost"
		host := "localhost"
		config := serviceConfig.GetBootstrap()

		if config.Service != nil {
			// These are needed so the Service can receive HTTP calls from services running in Docker, like Core Command to Device Service
			config.Service.Host = getLocalIP()
			config.Service.ServerBindAddr = "0.0.0.0"
		}

		if config.MessageBus != nil {
			config.MessageBus.Host = host
		}

		if config.Registry != nil {
			config.Registry.Host = host
		}

		if config.Database != nil {
			config.Database.Host = host
		}

		if config.Clients != nil {
			for _, client := range *config.Clients {
				client.Host = host
			}
		}
	}

	if remoteHosts != nil {
		err = applyRemoteHosts(remoteHosts, serviceConfig)
		if err != nil {
			return err
		}
	}

	return err
}

func (cp *Processor) loadConfigByProvider(
	configProviderInfo *ProviderInfo,
	configStem string,
	serviceConfig interfaces.Configuration,
	serviceType string,
	serviceKey string) (err error) {

	if err := cp.loadCommonConfig(configStem, configProviderInfo, serviceConfig, serviceType, CreateProviderClient); err != nil {
		return err
	}

	cp.lc.Info("Common configuration loaded from the Configuration Provider. No overrides applied")

	cp.privateConfigClient, err = CreateProviderClient(cp.lc, serviceKey, configStem, configProviderInfo.ServiceConfig())
	if err != nil {
		return fmt.Errorf("failed to create Configuration Provider client: %s", err.Error())
	}

	// TODO: figure out what uses the dic - this will not have the common config info!!
	// is this potentially custom config for app/device services?
	cp.dic.Update(di.ServiceConstructorMap{
		container.ConfigClientInterfaceName: func(get di.Get) any {
			return cp.privateConfigClient
		},
	})

	cp.providerHasConfig, err = cp.privateConfigClient.HasConfiguration()
	if err != nil {
		return fmt.Errorf("failed check for Configuration Provider has private configiuration: %s", err.Error())
	}

	if cp.providerHasConfig && !cp.getOverwriteConfig() {
		privateServiceConfig, err := copyConfigurationStruct(serviceConfig)
		if err != nil {
			return err
		}
		if err := cp.loadConfigFromProvider(privateServiceConfig, cp.privateConfigClient); err != nil {
			return err
		}
		configKeys, err := cp.privateConfigClient.GetConfigurationKeys("")
		if err != nil {
			return err
		}

		// Must remove any settings in the config that are not actually present in the Config Provider
		privateConfigKeys := utils.StringSliceToMap(configKeys)
		privateConfigMap, err := utils.RemoveUnusedSettings(privateServiceConfig, utils.BuildBaseKey(configStem, serviceKey), privateConfigKeys)
		if err != nil {
			return fmt.Errorf("could not remove unused settings from private configurations: %s", err.Error())
		}

		// Now merge only the actual present value with the existing configuration from common.
		if err := utils.MergeValues(serviceConfig, privateConfigMap); err != nil {
			return fmt.Errorf("could not merge common and private configurations: %s", err.Error())
		}

		cp.lc.Info("Private configuration loaded from the Configuration Provider. No overrides applied")
	} else {
		configMap, err := cp.loadPrivateConfigFromFile(serviceConfig)
		if err != nil {
			return err
		}

		if err := cp.privateConfigClient.PutConfigurationMap(configMap, cp.getOverwriteConfig()); err != nil {
			return fmt.Errorf("could not push private configuration into Configuration Provider: %s", err.Error())
		}

		cp.lc.Info("Private configuration has been pushed to into Configuration Provider with overrides applied")
	}

	cp.listenForPrivateChanges(serviceConfig, cp.privateConfigClient, utils.BuildBaseKey(configStem, serviceKey))
	cp.lc.Infof("listening for private config changes")
	cp.listenForCommonChanges(serviceConfig, cp.commonConfigClient, cp.privateConfigClient, utils.BuildBaseKey(configStem, common.CoreCommonConfigServiceKey, allServicesKey))
	cp.lc.Infof("listening for all services common config changes")
	if cp.appConfigClient != nil {
		cp.listenForCommonChanges(serviceConfig, cp.appConfigClient, cp.privateConfigClient, utils.BuildBaseKey(configStem, common.CoreCommonConfigServiceKey, appServicesKey))
		cp.lc.Infof("listening for application service common config changes")
	}
	if cp.deviceConfigClient != nil {
		cp.listenForCommonChanges(serviceConfig, cp.deviceConfigClient, cp.privateConfigClient, utils.BuildBaseKey(configStem, common.CoreCommonConfigServiceKey, deviceServicesKey))
		cp.lc.Infof("listening for device service common config changes")
	}
	return nil
}

func (cp *Processor) loadPrivateConfigFromFile(serviceConfig interfaces.Configuration) (configMap map[string]any, err error) {
	filePath := GetConfigFileLocation(cp.lc, cp.flags)
	configMap, err = cp.loadConfigYamlFromFile(filePath)

	if err != nil {
		return nil, err
	}

	// apply overrides - Now only done when loaded from file and values will get pushed into Configuration Provider (if used)
	overrideCount, err := cp.envVars.OverrideConfigMapValues(configMap)
	if err != nil {
		return nil, err
	}
	cp.lc.Infof("Private configuration loaded from file with %d overrides applied", overrideCount)

	if err := utils.MergeValues(serviceConfig, configMap); err != nil {
		return nil, err
	}

	return configMap, nil
}

func (cp *Processor) getOverwriteConfig() bool {
	// use the struct level pointer variable to prevent the logEnvironmentOverride log several times
	if cp.overwriteConfig != nil {
		return *cp.overwriteConfig
	}

	overwrite := cp.flags.OverwriteConfig()

	if b, ok := environment.OverwriteConfig(cp.lc); ok {
		overwrite = b
	}

	cp.overwriteConfig = &overwrite

	return overwrite
}

func getLocalIP() string {
	// Because of how UDP works, the connection is not established - no handshake is performed, no data is sent.
	// The purpose of this is to get the local IP address that a UDP connection would use if it were sending data to
	// the external destination address.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Since this is for Dev Mode, just default to localhost when can't get the actual IP
		return "localhost"
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP.String()
}

func applyRemoteHosts(remoteHosts []string, serviceConfig interfaces.Configuration) error {
	if len(remoteHosts) != 3 {
		return invalidRemoteHostsError
	}

	config := serviceConfig.GetBootstrap()

	config.Service.Host = remoteHosts[0]
	config.Service.ServerBindAddr = remoteHosts[2]

	if config.Config != nil {
		config.Config.Host = remoteHosts[1]
	}

	if config.MessageBus != nil {
		config.MessageBus.Host = remoteHosts[1]
	}

	if config.Registry != nil {
		config.Registry.Host = remoteHosts[1]
	}

	if config.Database != nil {
		config.Database.Host = remoteHosts[1]
	}

	if config.Clients != nil {
		for _, client := range *config.Clients {
			client.Host = remoteHosts[1]
		}
	}

	return nil
}

type createProviderCallback func(
	logger.LoggingClient,
	string,
	string,
	types.ServiceConfig) (configuration.Client, error)

// loadCommonConfig will pull up to two separate common configs from the config provider
// - serviceConfig: all services section of the common config
// - serviceTypeConfig: if the service is an app or device service, this will have the type specific common config
// if there are separate configs, these will get merged into the serviceConfig
func (cp *Processor) loadCommonConfig(
	configStem string,
	configProviderInfo *ProviderInfo,
	serviceConfig interfaces.Configuration,
	serviceType string,
	createProvider createProviderCallback) error {

	var err error
	// check that common config is loaded into the provider
	// this need a separate config provider client here because the config ready variable is stored at the common config level
	// load the all services section of the common config
	cp.commonConfigClient, err = createProvider(cp.lc, utils.BuildBaseKey(common.CoreCommonConfigServiceKey, allServicesKey), configStem, configProviderInfo.ServiceConfig())
	if err != nil {
		return fmt.Errorf("failed to create provider for %s: %s", allServicesKey, err.Error())
	}
	// build the path for the common configuration ready value
	commonConfigReadyPath := fmt.Sprintf("%s/%s/%s", configStem, common.CoreCommonConfigServiceKey, config.CommonConfigDone)
	if err := cp.waitForCommonConfig(cp.commonConfigClient, commonConfigReadyPath); err != nil {
		return err
	}
	err = cp.loadConfigFromProvider(serviceConfig, cp.commonConfigClient)
	if err != nil {
		return fmt.Errorf("failed to load the common configuration for %s: %s", allServicesKey, err.Error())
	}

	// use the service type to determine which additional sections to load into the common configuration
	var serviceTypeConfig interfaces.Configuration
	var serviceTypeConfigKeys []string
	var serviceTypeSectionKey string

	switch serviceType {
	case config.ServiceTypeApp:
		serviceTypeSectionKey = utils.BuildBaseKey(common.CoreCommonConfigServiceKey, appServicesKey)
		cp.lc.Infof("loading the common configuration for service type %s", serviceType)
		serviceTypeConfig, err = copyConfigurationStruct(serviceConfig)
		if err != nil {
			return fmt.Errorf("failed to copy the configuration structure for %s: %s", appServicesKey, err.Error())
		}
		cp.appConfigClient, err = createProvider(cp.lc, serviceTypeSectionKey, configStem, configProviderInfo.ServiceConfig())
		if err != nil {
			return fmt.Errorf("failed to create provider for %s: %s", appServicesKey, err.Error())
		}
		err = cp.loadConfigFromProvider(serviceTypeConfig, cp.appConfigClient)
		if err != nil {
			return fmt.Errorf("failed to load the common configuration for %s: %s", appServicesKey, err.Error())
		}
		serviceTypeConfigKeys, err = cp.appConfigClient.GetConfigurationKeys("")
		if err != nil {
			return fmt.Errorf("failed to load the common configuration keys for %s: %s", deviceServicesKey, err.Error())
		}

	case config.ServiceTypeDevice:
		serviceTypeSectionKey = utils.BuildBaseKey(common.CoreCommonConfigServiceKey, deviceServicesKey)
		cp.lc.Infof("loading the common configuration for service type %s", serviceType)
		serviceTypeConfig, err = copyConfigurationStruct(serviceConfig)
		if err != nil {
			return fmt.Errorf("failed to copy the configuration structure for %s: %s", deviceServicesKey, err.Error())
		}
		cp.deviceConfigClient, err = createProvider(cp.lc, serviceTypeSectionKey, configStem, configProviderInfo.ServiceConfig())
		if err != nil {
			return fmt.Errorf("failed to create provider for %s: %s", deviceServicesKey, err.Error())
		}
		err = cp.loadConfigFromProvider(serviceTypeConfig, cp.deviceConfigClient)
		if err != nil {
			return fmt.Errorf("failed to load the common configuration for %s: %s", deviceServicesKey, err.Error())
		}
		serviceTypeConfigKeys, err = cp.deviceConfigClient.GetConfigurationKeys("")
		if err != nil {
			return fmt.Errorf("failed to load the common configuration keys for %s: %s", deviceServicesKey, err.Error())
		}

	default:
		// this case is covered by the initial call to get the common config for all-services
	}

	// merge together the common config and the service type config
	if serviceTypeConfig != nil {
		// Must remove any settings in the config that are not actually present in the Config Provider
		serviceTypeConfigMap, err := utils.RemoveUnusedSettings(serviceTypeConfig, utils.BuildBaseKey(configStem, serviceTypeSectionKey), utils.StringSliceToMap(serviceTypeConfigKeys))
		if err != nil {
			return fmt.Errorf("failed to remove unused setting from %s common config: %s", serviceType, err.Error())
		}

		// merge common config and the service type common config's actually used settings
		if err := utils.MergeValues(serviceConfig, serviceTypeConfigMap); err != nil {
			return fmt.Errorf("failed to merge %s config with common config: %s", serviceType, err.Error())
		}
	}

	return nil
}

// loadCommonConfigFromFile will pull up the common config from the provided file and load it into the passed in interface
func (cp *Processor) loadCommonConfigFromFile(
	configFile string,
	serviceConfig interfaces.Configuration,
	serviceType string) error {

	var err error

	commonConfig, err := cp.loadConfigYamlFromFile(configFile)
	if err != nil {
		return err
	}
	// separate out the necessary sections
	allServicesConfig, ok := commonConfig[allServicesKey].(map[string]any)
	if !ok {
		return fmt.Errorf("could not find %s section in common config %s", allServicesKey, configFile)
	}
	// use the service type to separate out the necessary sections
	var serviceTypeConfig map[string]any
	switch serviceType {
	case config.ServiceTypeApp:
		cp.lc.Infof("loading the common configuration for service type %s", serviceType)
		serviceTypeConfig, ok = commonConfig[appServicesKey].(map[string]any)
		if !ok {
			return fmt.Errorf("could not find %s section in common config %s", appServicesKey, configFile)
		}
	case config.ServiceTypeDevice:
		cp.lc.Infof("loading the common configuration for service type %s", serviceType)
		serviceTypeConfig, ok = commonConfig[deviceServicesKey].(map[string]any)
		if !ok {
			return fmt.Errorf("could not find %s section in common config %s", deviceServicesKey, configFile)
		}
	default:
		// this case is covered by the initial call to get the common config for all-services
	}

	if serviceType == config.ServiceTypeApp || serviceType == config.ServiceTypeDevice {
		utils.MergeMaps(allServicesConfig, serviceTypeConfig)
	}

	if err := utils.ConvertFromMap(allServicesConfig, serviceConfig); err != nil {
		return fmt.Errorf("failed to convert common configuration into service's configuration: %v", err)
	}

	return err
}

// LoadCustomConfigSection loads the specified custom configuration section from file or Configuration provider.
// Section will be seed if Configuration provider does yet have it. This is used for structures custom configuration
// in App and Device services
func (cp *Processor) LoadCustomConfigSection(updatableConfig interfaces.UpdatableConfig, sectionName string) error {
	if cp.envVars == nil {
		cp.envVars = environment.NewVariables(cp.lc)
	}

	if cp.privateConfigClient == nil {
		cp.lc.Info("Skipping use of Configuration Provider for custom configuration: Provider not available")
		filePath := GetConfigFileLocation(cp.lc, cp.flags)
		configMap, err := cp.loadConfigYamlFromFile(filePath)
		if err != nil {
			return err
		}

		err = utils.ConvertFromMap(configMap, updatableConfig)
		if err != nil {
			return fmt.Errorf("failed to convert custom configuration into service's configuration: %v", err)
		}
	} else {
		cp.lc.Infof("Checking if custom configuration ('%s') exists in Configuration Provider", sectionName)

		exists, err := cp.privateConfigClient.HasSubConfiguration(sectionName)
		if err != nil {
			return fmt.Errorf(
				"unable to determine if custom configuration exists in Configuration Provider: %s",
				err.Error())
		}

		if exists && !cp.getOverwriteConfig() {
			rawConfig, err := cp.privateConfigClient.GetConfiguration(updatableConfig)
			if err != nil {
				return fmt.Errorf(
					"unable to get custom configuration from Configuration Provider: %s", err.Error())
			}

			err = utils.MergeValues(updatableConfig, rawConfig)
			if err != nil {
				return fmt.Errorf("unable to merge custom configuration from Configuration Provider")
			}

			cp.lc.Info("Loaded custom configuration from Configuration Provider, no overrides applied")
		} else {
			filePath := GetConfigFileLocation(cp.lc, cp.flags)
			configMap, err := cp.loadConfigYamlFromFile(filePath)
			if err != nil {
				return err
			}

			if err := utils.MergeValues(updatableConfig, configMap); err != nil {
				return err
			}

			// Must apply override before pushing into Configuration Provider
			overrideCount, err := cp.envVars.OverrideConfiguration(updatableConfig)
			if err != nil {
				return fmt.Errorf("unable to apply environment overrides: %s", err.Error())
			}

			cp.lc.Infof("Loaded custom configuration from File (%d envVars overrides applied)", overrideCount)

			mapToPush := make(map[string]any)
			err = utils.ConvertToMap(updatableConfig, &mapToPush)
			if err != nil {
				return err
			}

			err = cp.privateConfigClient.PutConfigurationMap(mapToPush, true)
			if err != nil {
				return fmt.Errorf("error pushing custom config to Configuration Provider: %s", err.Error())
			}

			var overwriteMessage = ""
			if exists && cp.getOverwriteConfig() {
				overwriteMessage = "(overwritten)"
			}
			cp.lc.Infof("Custom Config loaded from file and pushed to Configuration Provider %s", overwriteMessage)
		}
	}

	return nil
}

// ListenForCustomConfigChanges listens for changes to the specified custom configuration section. When changes occur it
// applies the changes to the custom configuration section and signals the changes have occurred.
func (cp *Processor) ListenForCustomConfigChanges(
	configToWatch any,
	sectionName string,
	changedCallback func(any)) {
	if cp.privateConfigClient == nil {
		cp.lc.Warnf("unable to watch custom configuration for changes: Configuration Provider not enabled")
		return
	}

	cp.wg.Add(1)
	go func() {
		defer cp.wg.Done()

		errorStream := make(chan error)
		defer close(errorStream)

		updateStream := make(chan any)
		defer close(updateStream)

		go cp.privateConfigClient.WatchForChanges(updateStream, errorStream, configToWatch, sectionName, cp.getMessageClient)

		isFirstUpdate := true

		for {
			select {
			case <-cp.ctx.Done():
				cp.privateConfigClient.StopWatching()
				cp.lc.Infof("Watching for '%s' configuration changes has stopped", sectionName)
				return

			case ex := <-errorStream:
				cp.lc.Error(ex.Error())

			case raw := <-updateStream:
				// Config Provider sends an update as soon as the watcher is connected even though there are not
				// any changes to the configuration. This causes an issue during start-up if there is an
				// envVars override of one of the Writable fields, so we must ignore the first update.
				if isFirstUpdate {
					isFirstUpdate = false
					continue
				}

				cp.lc.Infof("Updated custom configuration '%s' has been received from the Configuration Provider", sectionName)
				changedCallback(raw)
			}
		}
	}()

	cp.lc.Infof("Watching for custom configuration changes has started for `%s`", sectionName)
}

// CreateProviderClient creates and returns a configuration.Client instance and logs Client connection information
func CreateProviderClient(
	lc logger.LoggingClient,
	serviceKey string,
	configStem string,
	providerConfig types.ServiceConfig) (configuration.Client, error) {

	// The passed in configStem already contains the trailing '/' in most cases so must verify and add if missing.
	if configStem[len(configStem)-1] != '/' {
		configStem = configStem + "/"
	}

	// Note: Can't use filepath.Join as it uses `\` on Windows which Consul doesn't recognize as a path separator.
	providerConfig.BasePath = fmt.Sprintf("%s%s", configStem, serviceKey)

	lc.Info(fmt.Sprintf(
		"Using Configuration provider (%s) from: %s with base path of %s",
		providerConfig.Type,
		providerConfig.GetUrl(),
		providerConfig.BasePath))

	return configuration.NewConfigurationClient(providerConfig)
}

// loadConfigYamlFromFile attempts to read the specified configuration yaml file
func (cp *Processor) loadConfigYamlFromFile(yamlFile string) (map[string]any, error) {
	secretProvider := container.SecretProviderExtFrom(cp.dic.Get)

	cp.lc.Infof("Loading configuration file from %s", yamlFile)
	contents, err := file.Load(yamlFile, secretProvider, cp.lc)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %s", yamlFile, err.Error())
	}

	data := make(map[string]any)

	err = yaml.Unmarshal(contents, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall configuration file %s: %s", yamlFile, err.Error())
	}
	return data, nil
}

// GetConfigFileLocation uses the environment variables and flags to determine the location of the configuration
func GetConfigFileLocation(lc logger.LoggingClient, flags flags.Common) string {
	configFileName := environment.GetConfigFileName(lc, flags.ConfigFileName())

	// Check for uri path
	parsedUrl, err := url.Parse(configFileName)
	if err != nil {
		lc.Errorf("Could not parse file path: %v", err)
		return ""
	}

	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		return configFileName
	}

	configDir := environment.GetConfigDir(lc, flags.ConfigDirectory())
	profileDir := environment.GetProfileDir(lc, flags.Profile())
	return filepath.Join(configDir, profileDir, configFileName)
}

// listenForPrivateChanges leverages the Configuration Provider client's WatchForChanges() method to receive changes to and update the
// service's configuration writable sub-struct.  It's assumed the log level is universally part of the
// writable struct and this function explicitly updates the loggingClient's log level when new configuration changes
// are received.
func (cp *Processor) listenForPrivateChanges(serviceConfig interfaces.Configuration, privateConfigClient configuration.Client, baseKey string) {
	lc := cp.lc
	isFirstUpdate := true

	cp.wg.Add(1)
	go func() {
		defer cp.wg.Done()

		errorStream := make(chan error)
		defer close(errorStream)

		updateStream := make(chan any)
		defer close(updateStream)

		go privateConfigClient.WatchForChanges(updateStream, errorStream, serviceConfig.EmptyWritablePtr(), writableKey, cp.getMessageClient)

		for {
			select {
			case <-cp.ctx.Done():
				privateConfigClient.StopWatching()
				lc.Infof("Watching for '%s' configuration changes has stopped", writableKey)
				return

			case ex := <-errorStream:
				lc.Errorf("error occurred during listening to the configuration changes: %s", ex.Error())

			case raw, ok := <-updateStream:
				if !ok {
					return
				}

				usedKeys, err := privateConfigClient.GetConfigurationKeys(writableKey)
				if err != nil {
					lc.Errorf("failed to get list of private configuration keys for %s: %v", writableKey, err)
				}

				rawMap, err := utils.RemoveUnusedSettings(raw, utils.BuildBaseKey(baseKey, writableKey), utils.StringSliceToMap(usedKeys))
				if err != nil {
					lc.Errorf("failed to remove unused private settings in %s: %v", writableKey, err)
				}

				// Config Provider sends an update as soon as the watcher is connected even though there are not
				// any changes to the configuration. This causes an issue during start-up if there is an
				// envVars override of one of the Writable fields, so we must ignore the first update.
				if isFirstUpdate {
					isFirstUpdate = false
					continue
				}
				cp.applyWritableUpdates(serviceConfig, rawMap)
			}
		}
	}()
}

// listenForCommonChanges leverages the Configuration Provider client's WatchForChanges() method to receive changes to and update the
// service's common configuration writable sub-struct.
func (cp *Processor) listenForCommonChanges(fullServiceConfig interfaces.Configuration, configClient configuration.Client,
	privateConfigClient configuration.Client, baseKey string) {
	lc := cp.lc
	isFirstUpdate := true
	baseKey = utils.BuildBaseKey(baseKey, writableKey)

	cp.wg.Add(1)
	go func(fullServiceConfig interfaces.Configuration,
		commonConfigClient configuration.Client,
		privateConfigClient configuration.Client, baseKey string) {
		defer cp.wg.Done()

		var previousCommonWritable any

		errorStream := make(chan error)
		defer close(errorStream)

		updateStream := make(chan any)
		defer close(updateStream)

		go commonConfigClient.WatchForChanges(updateStream, errorStream, fullServiceConfig.EmptyWritablePtr(), writableKey, cp.getMessageClient)

		for {
			select {
			case <-cp.ctx.Done():
				commonConfigClient.StopWatching()
				lc.Infof("Watching for '%s' configuration changes has stopped", writableKey)
				return

			case ex := <-errorStream:
				lc.Errorf("error occurred during listening to the configuration changes: %s", ex.Error())

			case raw, ok := <-updateStream:
				if !ok {
					return
				}

				usedKeys, err := commonConfigClient.GetConfigurationKeys(writableKey)
				if err != nil {
					if err != nil {
						lc.Errorf("failed to get list of common configuration keys for %s: %v", writableKey, err)
					}
				}

				rawMap, err := utils.RemoveUnusedSettings(raw, baseKey, utils.StringSliceToMap(usedKeys))
				if err != nil {
					lc.Errorf("failed to remove unused common settings in %s: %v", writableKey, err)
				}

				// Config Provider sends an update as soon as the watcher is connected even though there are not
				// any changes to the configuration. This causes an issue during start-up if there is an
				// envVars override of one of the Writable fields, so on the first update we can just save a copy of the
				// common writable for comparison for future writable updates.
				if isFirstUpdate {
					isFirstUpdate = false
					previousCommonWritable = rawMap
					continue
				}

				if err := cp.processCommonConfigChange(fullServiceConfig, previousCommonWritable, rawMap, privateConfigClient, configClient); err != nil {
					lc.Error(err.Error())
				}

				// ensure that the local copy of the common writable gets updated no matter what
				previousCommonWritable = raw
			}
		}
	}(fullServiceConfig, configClient, privateConfigClient, baseKey)
}

func (cp *Processor) processCommonConfigChange(fullServiceConfig interfaces.Configuration, previousCommonWritable any, raw any, privateConfigClient configuration.Client, configClient configuration.Client) error {
	changedKey, found := cp.findChangedKey(previousCommonWritable, raw)
	if found {
		// Only need to check App/Device writable if change was made to the all-services writable
		if configClient == cp.commonConfigClient {
			// check if changed value (from all-services) is an App or Device common override
			otherConfigClient := cp.appConfigClient
			if otherConfigClient == nil {
				otherConfigClient = cp.deviceConfigClient
			}

			// This case should not happen at this point, but need to guard against nil pointer
			if otherConfigClient != nil {
				if cp.isKeyInConfig(configClient, changedKey) {
					cp.lc.Warnf("ignoring changed writable key %s overwritten in App or Device common writable", changedKey)
					return nil
				}
			}

		}

		// check if changed value is a private override
		if cp.isKeyInConfig(privateConfigClient, changedKey) {
			cp.lc.Warnf("ignoring changed writable key %s overwritten in private writable", changedKey)
			return nil
		}
	}

	cp.applyWritableUpdates(fullServiceConfig, raw)
	return nil
}

func (cp *Processor) findChangedKey(previous any, updated any) (string, bool) {
	var changedKey string
	var previousMap, updatedMap map[string]any
	if err := utils.ConvertToMap(previous, &previousMap); err != nil {
		cp.lc.Errorf("could not convert previous interface to map: %s", err.Error())
		return "", false
	}
	if err := utils.ConvertToMap(updated, &updatedMap); err != nil {
		cp.lc.Errorf("could not convert updated interface to map: %s", err.Error())
		return "", false
	}
	changedKey = walkMapForChange(previousMap, updatedMap, "")
	if changedKey == "" {
		// look the other way around to see if an item was removed
		changedKey = walkMapForChange(updatedMap, previousMap, "")
		if changedKey == "" {
			cp.lc.Error("could not find updated writable key or an error occurred")
			return "", false
		}
	}
	return changedKey, true
}

func (cp *Processor) applyWritableUpdates(serviceConfig interfaces.Configuration, raw any) {
	lc := cp.lc
	previousLogLevel := serviceConfig.GetLogLevel()
	previousTelemetryInterval := serviceConfig.GetTelemetryInfo().Interval

	var previousInsecureSecrets config.InsecureSecrets
	if err := utils.DeepCopy(serviceConfig.GetInsecureSecrets(), &previousInsecureSecrets); err != nil {
		lc.Errorf("failed to deep copy insecure secrets: %v", err)
	}

	if err := utils.MergeValues(serviceConfig.GetWritablePtr(), raw); err != nil {
		lc.Errorf("failed to apply Writable change to service configuration: %v", err)
	}

	currentInsecureSecrets := serviceConfig.GetInsecureSecrets()
	currentLogLevel := serviceConfig.GetLogLevel()
	currentTelemetryInterval := serviceConfig.GetTelemetryInfo().Interval

	lc.Info("Writable configuration has been updated from the Configuration Provider")

	// Note: Updates occur one setting at a time so only have to look for single changes
	switch {
	case currentLogLevel != previousLogLevel:
		_ = lc.SetLogLevel(serviceConfig.GetLogLevel())
		lc.Info(fmt.Sprintf("Logging level changed to %s", currentLogLevel))

	// InsecureSecrets (map) will be nil if not in the original TOML used to seed the Config Provider,
	// so ignore it if this is the case.
	case currentInsecureSecrets != nil &&
		!reflect.DeepEqual(currentInsecureSecrets, previousInsecureSecrets):
		lc.Info("Insecure Secrets have been updated")
		secretProvider := container.SecretProviderExtFrom(cp.dic.Get)
		if secretProvider != nil {
			// Find the updated secret's path and perform call backs.
			updatedSecrets := getSecretNamesChanged(previousInsecureSecrets, currentInsecureSecrets)
			for _, v := range updatedSecrets {
				secretProvider.SecretUpdatedAtSecretName(v)
			}
		}

	case currentTelemetryInterval != previousTelemetryInterval:
		lc.Info("Telemetry interval has been updated. Processing new value...")
		interval, err := time.ParseDuration(currentTelemetryInterval)
		if err != nil {
			lc.Errorf("update telemetry interval value is invalid time duration, using previous value: %s", err.Error())
			break
		}

		if interval == 0 {
			lc.Infof("0 specified for metrics reporting interval. Setting to max duration to effectively disable reporting.")
			interval = math.MaxInt64
		}

		metricsManager := container.MetricsManagerFrom(cp.dic.Get)
		if metricsManager == nil {
			lc.Error("metrics manager not available while updating telemetry interval")
			break
		}

		metricsManager.ResetInterval(interval)

	default:
		// Signal that configuration updates exists that have not already been processed.
		if cp.configUpdated != nil {
			cp.configUpdated <- struct{}{}
		}
	}
}

func (cp *Processor) waitForCommonConfig(configClient configuration.Client, configReadyPath string) error {
	// Wait for configuration provider to be available
	isAlive := false
	for cp.startupTimer.HasNotElapsed() {
		if configClient.IsAlive() {
			isAlive = true
			break
		}

		cp.lc.Warnf("Waiting for configuration provider to be available")

		select {
		case <-cp.ctx.Done():
			return errors.New("aborted waiting Configuration Provider to be available")
		default:
			cp.startupTimer.SleepForInterval()
			continue
		}
	}
	if !isAlive {
		return errors.New("configuration provider is not available")
	}

	// check to see if common config is loaded
	isConfigReady := false
	isCommonConfigReady := false
	for cp.startupTimer.HasNotElapsed() {
		commonConfigReady, err := configClient.GetConfigurationValueByFullPath(configReadyPath)
		if err != nil {
			cp.lc.Warn("waiting for Common Configuration to be available from config provider")
			cp.startupTimer.SleepForInterval()
			continue
		}

		isCommonConfigReady, err = strconv.ParseBool(string(commonConfigReady))
		if err != nil {
			cp.lc.Warnf("did not get boolean from config provider for %s: %s", configReadyPath, err.Error())
			isCommonConfigReady = false
		}
		if isCommonConfigReady {
			isConfigReady = true
			break
		}

		cp.lc.Warn("waiting for Common Configuration to be available from config provider")

		select {
		case <-cp.ctx.Done():
			return errors.New("aborted waiting for Common Configuration to be available")
		default:
			cp.startupTimer.SleepForInterval()
			continue
		}
	}
	if !isConfigReady {
		return errors.New("common config is not loaded - check to make sure core-common-config-bootstrapper ran")
	}
	return nil
}

// loadConfigFromProvider loads the config into the config structure
func (cp *Processor) loadConfigFromProvider(serviceConfig interfaces.Configuration, configClient configuration.Client) error {
	// pull common config and apply config to service config structure
	rawConfig, err := configClient.GetConfiguration(serviceConfig)
	if err != nil {
		return err
	}

	// update from raw
	ok := serviceConfig.UpdateFromRaw(rawConfig)
	if !ok {
		return fmt.Errorf("could not update service's configuration from raw")
	}

	return nil
}

// getMessageClient waits and gets the message client
func (cp *Processor) getMessageClient() (msgClient messaging.MessageClient) {
	// there's no startupTimer for cp created by NewProcessorForCustomConfig
	// add a new startupTimer here
	if !cp.startupTimer.HasNotElapsed() {
		cp.startupTimer = startup.NewStartUpTimer("")
	}
	for cp.startupTimer.HasNotElapsed() {
		if msgClient = container.MessagingClientFrom(cp.dic.Get); msgClient != nil {
			break
		}
		cp.startupTimer.SleepForInterval()
	}
	if msgClient == nil {
		cp.lc.Error("unable to use MessageClient to watch for configuration changes")
		return nil
	}
	return msgClient
}

// getSecretNamesChanged returns a slice of secretNames that have changed secrets or are new.
func getSecretNamesChanged(prevVals config.InsecureSecrets, curVals config.InsecureSecrets) []string {
	var updatedNames []string
	for key, prevVal := range prevVals {
		curVal := curVals[key]

		// Catches removed secrets
		if curVal.SecretData == nil {
			updatedNames = append(updatedNames, prevVal.SecretName)
			continue
		}

		// Catches changes to secret data or to the secret name
		if !reflect.DeepEqual(prevVal, curVal) {
			updatedNames = append(updatedNames, curVal.SecretName)

			// Catches secret name changes, so also include the previous secretName
			if prevVal.SecretName != curVal.SecretName {
				updatedNames = append(updatedNames, prevVal.SecretName)
			}
		}
	}

	for key, curVal := range curVals {
		// Catches new secrets added
		if prevVals[key].SecretData == nil {
			updatedNames = append(updatedNames, curVal.SecretName)
		}
	}

	return updatedNames
}

// copyConfigurationStruct returns a copy of the passed in configuration interface
func copyConfigurationStruct(config interfaces.Configuration) (interfaces.Configuration, error) {
	rawCopy, err := copystructure.Copy(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load copy the configuration: %s", err.Error())
	}
	configCopy, ok := rawCopy.(interfaces.Configuration)
	if !ok {
		return nil, errors.New("failed to cast the copy of the configuration")
	}
	return configCopy, nil
}

func walkMapForChange(previousMap map[string]any, updatedMap map[string]any, changedKey string) string {
	for updatedKey, updatedVal := range updatedMap {
		previousVal, ok := previousMap[updatedKey]
		if !ok {
			return buildNewKey(changedKey, updatedKey)
		}
		updatedSubMap, ok := updatedVal.(map[string]any)
		// if the value is not of type map[string]any, it should be a value to compare
		if !ok {
			if updatedVal != previousVal {
				return buildNewKey(changedKey, updatedKey)
			}
			continue
		}
		previousSubMap, ok := previousVal.(map[string]any)
		if !ok {
			// handle the case where a new setting is added
			if previousSubMap == nil && updatedSubMap != nil {
				subKey := buildNewKey(changedKey, updatedKey)
				for k := range updatedSubMap {
					return buildNewKey(subKey, k)
				}
			}
			return ""
		}
		key := buildNewKey(changedKey, updatedKey)
		key = walkMapForChange(previousSubMap, updatedSubMap, key)
		if len(key) > 0 {
			return key
		}
	}
	return ""
}

func (cp *Processor) isKeyInConfig(configClient configuration.Client, changedKey string) bool {
	keys, err := configClient.GetConfigurationKeys(writableKey)
	if err != nil {
		cp.lc.Errorf("could not get writable keys from configuration: %s", err.Error())
		// return true because shouldn't change an overridden value
		// error means it is undetermined, so don't override to be safe
		return true
	}
	changedKey = fmt.Sprintf("%s/%s", writableKey, changedKey)

	for _, key := range keys {
		if strings.Contains(key, changedKey) {
			return true
		}
	}
	return false
}

func buildNewKey(previousKey, currentKey string) string {
	if previousKey != "" {
		return utils.BuildBaseKey(previousKey, currentKey)
	} else {
		return currentKey
	}
}

// GetInsecureSecretNameFullPath returns the full configuration path of an InsecureSecret's SecretName field.
// example: Writable/InsecureSecrets/credentials001/SecretName
func GetInsecureSecretNameFullPath(secretName string) string {
	return fmt.Sprintf("%s/%s/%s/%s",
		writableKey, insecureSecretsKey, secretName, secretNameKey)
}

// GetInsecureSecretDataFullPath returns the full configuration path of an InsecureSecret's SecretData for a specific key.
// example: Writable/InsecureSecrets/credentials001/SecretData/username
func GetInsecureSecretDataFullPath(secretName, key string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s",
		writableKey, insecureSecretsKey, secretName, secretDataKey, key)
}
