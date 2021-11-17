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
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/edgexfoundry/go-mod-configuration/v2/pkg/types"

	"github.com/pelletier/go-toml"
)

const (
	bootTimeoutSecondsDefault = 60
	bootRetrySecondsDefault   = 1
	defaultConfDirValue       = "./res"

	envKeyConfigUrl       = "EDGEX_CONFIGURATION_PROVIDER"
	envKeyUseRegistry     = "EDGEX_USE_REGISTRY"
	envKeyStartupDuration = "EDGEX_STARTUP_DURATION"
	envKeyStartupInterval = "EDGEX_STARTUP_INTERVAL"
	envConfDir            = "EDGEX_CONF_DIR"
	envProfile            = "EDGEX_PROFILE"
	envFile               = "EDGEX_CONFIG_FILE"

	tomlPathSeparator = "."
	tomlNameSeparator = "-"
	envNameSeparator  = "_"
)

// Variables is receiver that holds Variables variables and encapsulates toml.Tree-based configuration field
// overrides.  Assumes "_" embedded in Variables variable key separates sub-structs; e.g. foo_bar_baz might refer to
//
// 		type foo struct {
// 			bar struct {
//          	baz string
//  		}
//		}
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
func (e *Variables) OverrideConfiguration(serviceConfig interface{}) (int, error) {
	var overrideCount = 0

	contents, err := toml.Marshal(reflect.ValueOf(serviceConfig).Elem().Interface())
	if err != nil {
		return 0, err
	}

	configTree, err := toml.LoadBytes(contents)
	if err != nil {
		return 0, err
	}

	// The toml.Tree API keys() only return to top level keys, rather that paths.
	// It is also missing a GetPaths so have to spin our own
	paths := e.buildPaths(configTree.ToMap())
	// Now that we have all the paths in the config tree, we need to create map of corresponding override names that
	// could match override environment variable names.
	overrideNames := e.buildOverrideNames(paths)

	for envVar, envValue := range e.variables {
		path, found := overrideNames[envVar]
		if !found {
			continue
		}

		oldValue := configTree.Get(path)

		newValue, err := e.convertToType(oldValue, envValue)
		if err != nil {
			return 0, fmt.Errorf("environment value override failed for %s=%s: %s", envVar, envValue, err.Error())
		}

		configTree.Set(path, newValue)
		overrideCount++
		logEnvironmentOverride(e.lc, path, envVar, envValue)
	}

	// Put the configuration back into the services configuration struct with the overridden values
	err = configTree.Unmarshal(serviceConfig)
	if err != nil {
		return 0, fmt.Errorf("could not marshal toml configTree to configuration: %s", err.Error())
	}

	return overrideCount, nil
}

// buildPaths create the path strings for all settings in the Config tree's key map
func (e *Variables) buildPaths(keyMap map[string]interface{}) []string {
	var paths []string

	for key, item := range keyMap {
		if reflect.TypeOf(item).Kind() != reflect.Map {
			paths = append(paths, key)
			continue
		}

		subMap := item.(map[string]interface{})

		subPaths := e.buildPaths(subMap)
		for _, path := range subPaths {
			paths = append(paths, fmt.Sprintf("%s.%s", key, path))
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
	// "." & "-" are the only special character allowed in TOML path not allowed in environment variable Name
	override := strings.ReplaceAll(path, tomlPathSeparator, envNameSeparator)
	override = strings.ReplaceAll(override, tomlNameSeparator, envNameSeparator)
	override = strings.ToUpper(override)
	return override
}

// OverrideConfigProviderInfo overrides the Configuration Provider ServiceConfig values
// from an Variables variable value (if it exists).
func (e *Variables) OverrideConfigProviderInfo(configProviderInfo types.ServiceConfig) (types.ServiceConfig, error) {

	url := os.Getenv(envKeyConfigUrl)
	if len(url) > 0 {
		logEnvironmentOverride(e.lc, "Configuration Provider Information", envKeyConfigUrl, url)

		if err := configProviderInfo.PopulateFromUrl(url); err != nil {
			return types.ServiceConfig{}, err
		}
	}

	return configProviderInfo, nil
}

// convertToType attempts to convert the string value to the specified type of the old value
func (_ *Variables) convertToType(oldValue interface{}, value string) (newValue interface{}, err error) {
	switch oldValue.(type) {
	case []string:
		newValue = parseCommaSeparatedSlice(value)
	case []interface{}:
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
	// lc hasn't be created at the time this info is needed so have to create local client.
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

// GetConfDir get the config directory value from an Variables variable value (if it exists)
// or uses passed in value or default if previous result in blank.
func GetConfDir(lc logger.LoggingClient, configDir string) string {
	envValue := os.Getenv(envConfDir)
	if len(envValue) > 0 {
		configDir = envValue
		logEnvironmentOverride(lc, "-c/-confdir", envConfDir, envValue)
	}

	if len(configDir) == 0 {
		configDir = defaultConfDirValue
	}

	return configDir
}

// GetProfileDir get the profile directory value from an Variables variable value (if it exists)
// or uses passed in value or default if previous result in blank.
func GetProfileDir(lc logger.LoggingClient, profileDir string) string {
	envValue := os.Getenv(envProfile)
	if len(envValue) > 0 {
		profileDir = envValue
		logEnvironmentOverride(lc, "-p/-profile", envProfile, envValue)
	}

	if len(profileDir) > 0 {
		profileDir += "/"
	}

	return profileDir
}

// GetConfigFileName gets the configuration filename value from an Variables variable value (if it exists)
// or uses passed in value.
func GetConfigFileName(lc logger.LoggingClient, configFileName string) string {
	envValue := os.Getenv(envFile)
	if len(envValue) > 0 {
		configFileName = envValue
		logEnvironmentOverride(lc, "-f/-file", envFile, envValue)
	}

	return configFileName
}

// parseCommaSeparatedSlice converts comma separated list to a string slice
func parseCommaSeparatedSlice(value string) (values []interface{}) {
	// Assumption is environment variable value is comma separated
	// Whitespace can vary so must be trimmed out
	result := strings.Split(strings.TrimSpace(value), ",")
	for _, entry := range result {
		values = append(values, strings.TrimSpace(entry))
	}

	return values
}

// logEnvironmentOverride logs that an option or configuration has been override by an environment variable.
func logEnvironmentOverride(lc logger.LoggingClient, name string, key string, value string) {
	lc.Info(fmt.Sprintf("Variables override of '%s' by environment variable: %s=%s", name, key, value))
}
