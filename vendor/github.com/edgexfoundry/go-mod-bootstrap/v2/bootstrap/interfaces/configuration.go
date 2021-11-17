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

package interfaces

import "github.com/edgexfoundry/go-mod-bootstrap/v2/config"

// UpdatableConfig interface allows service to have their custom configuration populated from configuration stored
// in the Configuration Provider (aka Consul). A service using custom configuration must implement this interface
// on the custom configuration, even if not using Configuration Provider. If not using the Configuration Provider
// it can have dummy implementations of this interface.
type UpdatableConfig interface {
	// UpdateFromRaw converts configuration received from the Configuration Provider to a service-specific
	// configuration struct which is then used to overwrite the service's existing configuration struct.
	UpdateFromRaw(rawConfig interface{}) bool
}

// WritableConfig allows service to listen for changes from the Configuration Provider and have the configuration updated
// when the changes occur
type WritableConfig interface {
	// UpdateWritableFromRaw converts updated configuration received from the Configuration Provider to a
	// service-specific struct that is being watched for changes by the Configuration Provider.
	// The changes are used to overwrite the service's existing configuration's watched struct.
	UpdateWritableFromRaw(rawWritableConfig interface{}) bool
}

// Configuration interface provides an abstraction around a configuration struct.
type Configuration interface {
	// These two interfaces have been separated out for use in the custom configuration capability for
	// App and Device services
	UpdatableConfig
	WritableConfig

	// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.  It is used by the bootstrap to
	// provide the appropriate structure to registry.Client's WatchForChanges().
	EmptyWritablePtr() interface{}

	// GetBootstrap returns the configuration elements required by the bootstrap.
	GetBootstrap() config.BootstrapConfiguration

	// GetLogLevel returns the current ConfigurationStruct's log level.
	GetLogLevel() string

	// GetRegistryInfo gets the config.RegistryInfo field from the ConfigurationStruct.
	GetRegistryInfo() config.RegistryInfo

	// GetInsecureSecrets gets the config.InsecureSecrets field from the ConfigurationStruct.
	GetInsecureSecrets() config.InsecureSecrets
}
