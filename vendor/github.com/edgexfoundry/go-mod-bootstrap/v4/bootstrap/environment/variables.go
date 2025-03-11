/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2020 Intel Inc.
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

package environment

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/go-mod-configuration/v4/pkg/types"
)

const (
	bootTimeoutSecondsDefault = 60
	bootRetrySecondsDefault   = 1
	defaultConfigDirValue     = "./res"

	envKeyConfigUrl          = "EDGEX_CONFIG_PROVIDER"
	envKeyCommonConfig       = "EDGEX_COMMON_CONFIG"
	envKeyUseRegistry        = "EDGEX_USE_REGISTRY"
	envKeyStartupDuration    = "EDGEX_STARTUP_DURATION"
	envKeyStartupInterval    = "EDGEX_STARTUP_INTERVAL"
	envKeyConfigDir          = "EDGEX_CONFIG_DIR"
	envKeyProfile            = "EDGEX_PROFILE"
	envKeyConfigFile         = "EDGEX_CONFIG_FILE"
	envKeyFileURITimeout     = "EDGEX_FILE_URI_TIMEOUT"
	envKeyRemoteServiceHosts = "EDGEX_REMOTE_SERVICE_HOSTS"
	envKeyOverwriteConfig    = "EDGEX_OVERWRITE_CONFIG"

	noConfigProviderValue = "none"

	configPathSeparator = "/"
	configNameSeparator = "-"
	envNameSeparator    = "_"

	// insecureSecretsRegexStr is a regex to look for toml keys that are under the Secrets sub-key of values within the
	// Writable.InsecureSecrets topology.
	// Examples:
	//			Matches: Writable.InsecureSecrets.credentials001.Secrets.password
	//	 Does Not Match: Writable.InsecureSecrets.credentials001.Path
	insecureSecretsRegexStr = "^Writable\\.InsecureSecrets\\.[^.]+\\.Secrets\\..+$" //#nosec G101 -- This is a false positive
	// redactedStr is the value to print for redacted variable values
	redactedStr = "<redacted>"

	defaultFileURITimeout = 15 * time.Second
)

var (
	insecureSecretsRegex = regexp.MustCompile(insecureSecretsRegexStr)
)

// Variables is a receiver that holds Variables and encapsulates toml.Tree-based configuration field
// overrides.  Assumes "_" embedded in Variables variable key separates sub-structs; e.g. foo_bar_baz might refer to
//
//			type foo struct {
//				bar struct {
//	         		baz string
//	 			}
//			}
type Variables struct {
	variables map[string]string
	lc        logger.LoggingClient
}

// NewVariables constructor reads/stores os.Environ() for use by Variables receiver methods.
func NewVariables(lc logger.LoggingClient) *Variables {
	osEnv := os.Environ()
	e := &Variables{
		variables: make(map[string]string, len(osEnv)),
		lc:        lc,
	}

	for _, env := range osEnv {
		// Can not use Split() on '=' since the value may have an '=' in it, so changed to use Index()
		index := strings.Index(env, "=")
		if index == -1 {
			continue
		}
		key := env[:index]
		value := env[index+1:]
		e.variables[key] = value
	}

	return e
}

// UseRegistry returns whether the envKeyUseRegistry key is set to true and whether the override was used
func (e *Variables) UseRegistry() (bool, bool) {
	value := os.Getenv(envKeyUseRegistry)
	if len(value) == 0 {
		return false, false
	}

	logEnvironmentOverride(e.lc, "-r/--registry", envKeyUseRegistry, value)

	return value == "true", true
}

// OverrideConfiguration method replaces values in the configuration for matching Variables variable keys.
// serviceConfig must be pointer to the service configuration.
func (e *Variables) OverrideConfiguration(serviceConfig any) (int, error) {

	contents, err := json.Marshal(reflect.ValueOf(serviceConfig).Elem().Interface())
	if err != nil {
		return 0, err
	}

	configMap := make(map[string]any)
	err = json.Unmarshal(contents, &configMap)
	if err != nil {
		return 0, err
	}

	overrideCount, err := e.OverrideConfigMapValues(configMap)
	if err != nil {
		return 0, err
	}

	// Put the configuration back into the services configuration struct with the overridden values
	err = utils.ConvertFromMap(configMap, serviceConfig)
	if err != nil {
		return 0, fmt.Errorf("failed to convert map of configuratuion into service's configuration struct: %v", err)
	}

	return overrideCount, nil
}

func (e *Variables) OverrideConfigMapValues(configMap map[string]any) (int, error) {
	var overrideCount int

	// The toml.Tree API keys() only return to top level keys, rather that paths.
	// It is also missing a GetPaths so have to spin our own
	paths := e.buildPaths(configMap)
	// Now that we have all the paths in the config tree, we need to create map of corresponding override names that
	// could match override environment variable names.
	overrideNames := e.buildOverrideNames(paths)

	for envVar, envValue := range e.variables {
		path, found := overrideNames[envVar]
		if !found {
			continue
		}

		oldValue := getConfigMapValue(path, configMap)

		newValue, err := e.convertToType(oldValue, envValue)
		if err != nil {
			return 0, fmt.Errorf("environment value override failed for %s=%s: %s", envVar, envValue, err.Error())
		}

		setConfigMapValue(path, newValue, configMap)
		overrideCount++
		logEnvironmentOverride(e.lc, path, envVar, envValue)
	}

	return overrideCount, nil
}

func getConfigMapValue(path string, configMap map[string]any) any {
	// First check the case of flattened map where the path is the key
	value, exists := configMap[path]
	if exists {
		return value
	}

	// Handle second case of not flattened map where the path is individual keys
	keys := strings.Split(path, configPathSeparator)

	currentMap := configMap

	for _, key := range keys {
		item := currentMap[key]
		if item == nil {
			return nil
		}

		itemMap, isMap := item.(map[string]any)
		if !isMap {
			return item
		}

		currentMap = itemMap
		continue
	}

	return nil
}

func setConfigMapValue(path string, value any, configMap map[string]any) {
	// First check the case of flattened map where the path is the key
	_, exists := configMap[path]
	if exists {
		configMap[path] = value
		return
	}

	// Handle second case of not flattened map where the path is individual keys
	keys := strings.Split(path, configPathSeparator)

	currentMap := configMap

	for _, key := range keys {
		item := currentMap[key]
		itemMap, isMap := item.(map[string]any)
		if !isMap {
			currentMap[key] = value
			return
		}

		currentMap = itemMap
		continue
	}
}

// buildPaths create the path strings for all settings in the Config key map
func (e *Variables) buildPaths(keyMap map[string]any) []string {
	var paths []string

	for key, item := range keyMap {
		if item == nil || reflect.TypeOf(item).Kind() != reflect.Map {
			paths = append(paths, key)
			continue
		}

		subMap := item.(map[string]any)

		subPaths := e.buildPaths(subMap)
		for _, path := range subPaths {
			paths = append(paths, fmt.Sprintf("%s/%s", key, path))
		}
	}

	return paths
}

func (e *Variables) buildOverrideNames(paths []string) map[string]string {
	names := map[string]string{}
	for _, path := range paths {
		names[e.getOverrideNameFor(path)] = path
	}

	return names
}

func (_ *Variables) getOverrideNameFor(path string) string {
	// "/" & "-" are the only special character allowed in path not allowed in environment variable Name
	override := strings.ReplaceAll(path, configPathSeparator, envNameSeparator)
	override = strings.ReplaceAll(override, configNameSeparator, envNameSeparator)
	override = strings.ToUpper(override)
	return override
}

// OverrideConfigProviderInfo overrides the Configuration Provider ServiceConfig values
// from an Variables variable value (if it exists).
func (e *Variables) OverrideConfigProviderInfo(configProviderInfo types.ServiceConfig) (types.ServiceConfig, error) {

	url := os.Getenv(envKeyConfigUrl)
	if len(url) > 0 {
		logEnvironmentOverride(e.lc, "Configuration Provider Information", envKeyConfigUrl, url)

		if url == noConfigProviderValue {
			return types.ServiceConfig{}, nil
		}

		if err := configProviderInfo.PopulateFromUrl(url); err != nil {
			return types.ServiceConfig{}, err
		}
	}

	return configProviderInfo, nil
}

// convertToType attempts to convert the string value to the specified type of the old value
func (_ *Variables) convertToType(oldValue any, value string) (newValue any, err error) {
	switch oldValue.(type) {
	case []string:
		newValue = parseCommaSeparatedSlice(value)
	case []any:
		newValue = parseCommaSeparatedSlice(value)
	case string:
		newValue = value
	case bool:
		newValue, err = strconv.ParseBool(value)
	case int:
		newValue, err = strconv.ParseInt(value, 10, strconv.IntSize)
		newValue = int(newValue.(int64))
	case int8:
		newValue, err = strconv.ParseInt(value, 10, 8)
		newValue = int8(newValue.(int64))
	case int16:
		newValue, err = strconv.ParseInt(value, 10, 16)
		newValue = int16(newValue.(int64))
	case int32:
		newValue, err = strconv.ParseInt(value, 10, 32)
		newValue = int32(newValue.(int64))
	case int64:
		newValue, err = strconv.ParseInt(value, 10, 64)
	case uint:
		newValue, err = strconv.ParseUint(value, 10, strconv.IntSize)
		newValue = uint(newValue.(uint64))
	case uint8:
		newValue, err = strconv.ParseUint(value, 10, 8)
		newValue = uint8(newValue.(uint64))
	case uint16:
		newValue, err = strconv.ParseUint(value, 10, 16)
		newValue = uint16(newValue.(uint64))
	case uint32:
		newValue, err = strconv.ParseUint(value, 10, 32)
		newValue = uint32(newValue.(uint64))
	case uint64:
		newValue, err = strconv.ParseUint(value, 10, 64)
	case float32:
		newValue, err = strconv.ParseFloat(value, 32)
		newValue = float32(newValue.(float64))
	case float64:
		newValue, err = strconv.ParseFloat(value, 64)
	default:
		err = fmt.Errorf(
			"configuration type of '%s' is not supported for environment variable override",
			reflect.TypeOf(oldValue).String())
	}

	return newValue, err
}

// StartupInfo provides the startup timer values which are applied to the StartupTimer created at boot.
type StartupInfo struct {
	Duration int
	Interval int
}

// GetStartupInfo gets the Service StartupInfo values from an Variables variable value (if it exists)
// or uses the default values.
func GetStartupInfo(serviceKey string) StartupInfo {
	// lc hasn't been created at the time this info is needed so have to create local client.
	lc := logger.NewClient(serviceKey, models.InfoLog)

	startup := StartupInfo{
		Duration: bootTimeoutSecondsDefault,
		Interval: bootRetrySecondsDefault,
	}

	// Get the startup timer configuration from environment, if provided.
	value := os.Getenv(envKeyStartupDuration)
	if len(value) > 0 {
		logEnvironmentOverride(lc, "Startup Duration", envKeyStartupDuration, value)

		if n, err := strconv.ParseInt(value, 10, 0); err == nil && n > 0 {
			startup.Duration = int(n)
		}
	}

	// Get the startup timer interval, if provided.
	value = os.Getenv(envKeyStartupInterval)
	if len(value) > 0 {
		logEnvironmentOverride(lc, "Startup Interval", envKeyStartupInterval, value)

		if n, err := strconv.ParseInt(value, 10, 0); err == nil && n > 0 {
			startup.Interval = int(n)
		}
	}

	return startup
}

// GetConfigDir get the config directory value from a Variables variable value (if it exists)
// or uses passed in value or default if previous result in blank.
func GetConfigDir(lc logger.LoggingClient, configDir string) string {
	envValue := os.Getenv(envKeyConfigDir)
	if len(envValue) > 0 {
		configDir = envValue
		logEnvironmentOverride(lc, "-cd/-configDir", envKeyConfigDir, envValue)
	}

	if len(configDir) == 0 {
		configDir = defaultConfigDirValue
	}

	return configDir
}

// GetProfileDir get the profile directory value from a Variables variable value (if it exists)
// or uses passed in value or default if previous result in blank.
func GetProfileDir(lc logger.LoggingClient, profileDir string) string {
	envValue := os.Getenv(envKeyProfile)
	if len(envValue) > 0 {
		profileDir = envValue
		logEnvironmentOverride(lc, "-p/-profile", envKeyProfile, envValue)
	}

	if len(profileDir) > 0 {
		profileDir += "/"
	}

	return profileDir
}

// GetConfigFileName gets the configuration filename value from a Variables variable value (if it exists)
// or uses passed in value.
func GetConfigFileName(lc logger.LoggingClient, configFileName string) string {
	envValue := os.Getenv(envKeyConfigFile)
	if len(envValue) > 0 {
		configFileName = envValue
		logEnvironmentOverride(lc, "-cf/--configFile", envKeyConfigFile, envValue)
	}

	return configFileName
}

// GetCommonConfigFileName gets the common configuration value from the Variables value (if it exists)
// or uses passed in value.
func GetCommonConfigFileName(lc logger.LoggingClient, commonConfigFileName string) string {
	envValue := os.Getenv(envKeyCommonConfig)
	if len(envValue) > 0 {
		commonConfigFileName = envValue
		logEnvironmentOverride(lc, "-cc/--commonConfig", envKeyCommonConfig, envValue)
	}

	return commonConfigFileName
}

// parseCommaSeparatedSlice converts comma separated list to a string slice
func parseCommaSeparatedSlice(value string) (values []any) {
	// Assumption is environment variable value is comma separated
	// Whitespace can vary so must be trimmed out
	result := strings.Split(strings.TrimSpace(value), ",")
	for _, entry := range result {
		values = append(values, strings.TrimSpace(entry))
	}

	return values
}

// logEnvironmentOverride logs that an option or configuration has been override by an environment variable.
// If the key belongs to a Secret within Writable.InsecureSecrets, the value is redacted when printing it.
func logEnvironmentOverride(lc logger.LoggingClient, name string, key string, value string) {
	valueStr := value
	if insecureSecretsRegex.MatchString(name) {
		valueStr = redactedStr
	}
	lc.Infof("Variables override of '%s' by environment variable: %s=%s", name, key, valueStr)
}

// GetURIRequestTimeout gets the configuration request timeout value from an environment variable (if it exists)
// or uses default value.
func GetURIRequestTimeout(lc logger.LoggingClient) time.Duration {
	envValue := os.Getenv(envKeyFileURITimeout)
	if len(envValue) <= 0 {
		return defaultFileURITimeout
	}

	requestTimeout, err := time.ParseDuration(envValue)
	if err != nil {
		lc.Warnf("Could not parse value for %s = %s: %v. Using default of %s", envKeyFileURITimeout, envValue, err, defaultFileURITimeout)
		return defaultFileURITimeout
	}

	lc.Infof("Variables override of 'URI Request Timeout' by environment variable: %s=%s", envKeyFileURITimeout, envValue)
	return requestTimeout
}

// GetRemoteServiceHosts gets the Remote Service host name list from an environment variable (if it exists), if not returns the passed in (default) value
func GetRemoteServiceHosts(lc logger.LoggingClient, remoteHosts []string) []string {
	envValue := os.Getenv(envKeyRemoteServiceHosts)
	if len(envValue) <= 0 {
		return remoteHosts
	}

	logEnvironmentOverride(lc, "-rsh/--remoteServiceHosts", envKeyRemoteServiceHosts, envValue)

	return strings.Split(envValue, ",")
}

// OverwriteConfig returns whether the local configuration should be pushed (overwrite) into the Configuration provider.
// The second bool return value indicates whether the environment variable is set.
func OverwriteConfig(lc logger.LoggingClient) (bool, bool) {
	envValue := os.Getenv(envKeyOverwriteConfig)
	if len(envValue) == 0 {
		return false, false
	}

	logEnvironmentOverride(lc, "-o/--overwrite", envKeyOverwriteConfig, envValue)

	boolValue, err := strconv.ParseBool(envValue)
	if err != nil {
		return false, true
	}

	return boolValue, true
}
