//
// Copyright (c) 2019 Intel Corporation
// Copyright (C) 2024-2025 IOTech Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package configuration

import "github.com/edgexfoundry/go-mod-messaging/v4/messaging"

type Client interface {
	// HasConfiguration checks to see if the Configuration service contains the service's configuration.
	HasConfiguration() (bool, error)

	// HasSubConfiguration checks to see if the Configuration service contains the service's sub configuration.
	HasSubConfiguration(name string) (bool, error)

	// PutConfigurationMap puts a full map configuration into the Configuration service
	// The sub-paths to where the values are to be stored in the Configuration service are generated from the map key.
	PutConfigurationMap(configuration map[string]any, overwrite bool) error

	// PutConfiguration puts a full configuration struct into the Configuration service
	PutConfiguration(configStruct interface{}, overwrite bool) error

	// GetConfiguration gets the full configuration from keeper into the target configuration struct.
	// Passed in struct is only a reference for Configuration service. Empty struct is fine
	// Returns the configuration in the target struct as interface{}, which caller must cast
	GetConfiguration(configStruct interface{}) (interface{}, error)

	// WatchForChanges sets up a keeper watch for the target key and send back updates on the update channel.
	// Passed in struct is only a reference for Configuration service, empty struct is ok
	// Sends the configuration in the target struct as interface{} on updateChannel, which caller must cast
	WatchForChanges(updateChannel chan<- interface{}, errorChannel chan<- error, configuration interface{}, waitKey string, getMsgClientCb func() messaging.MessageClient)

	// StopWatching causes all WatchForChanges processing to stop and waits until they have stopped.
	StopWatching()

	// IsAlive simply checks if Configuration service is up and running at the configured URL
	IsAlive() bool

	// ConfigurationValueExists checks if a configuration value exists in the Configuration service
	ConfigurationValueExists(name string) (bool, error)

	// GetConfigurationValue gets a specific configuration value from the Configuration service
	GetConfigurationValue(name string) ([]byte, error)

	// GetConfigurationValueByFullPath gets a specific configuration value from the Configuration service
	GetConfigurationValueByFullPath(fullPath string) ([]byte, error)

	// PutConfigurationValue puts a specific configuration value into the Configuration service
	PutConfigurationValue(name string, value []byte) error

	// GetConfigurationKeys returns all keys under name
	GetConfigurationKeys(name string) ([]string, error)
}
