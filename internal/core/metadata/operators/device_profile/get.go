/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package device_profile

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// GetProfilesExecutor retrieves one or more device profiles based on some criteria.
type GetProfilesExecutor interface {
	Execute() ([]contract.DeviceProfile, error)
}

// GetProfilesExecutor retrieves a profile based on some criteria.
type GetProfileExecutor interface {
	Execute() (contract.DeviceProfile, error)
}

// getAllDeviceProfiles encapsulates the data needed in order to get all device profiles.
type getAllDeviceProfiles struct {
	config bootstrapConfig.ServiceInfo
	loader DeviceProfileLoader
	logger logger.LoggingClient
}

// Execute retrieves all device profiles.
func (g getAllDeviceProfiles) Execute() ([]contract.DeviceProfile, error) {
	dps, err := g.loader.GetAllDeviceProfiles()
	if err != nil {
		return nil, err
	}

	if len(dps) > g.config.MaxResultCount {
		err = errors.NewErrLimitExceeded(g.config.MaxResultCount)
		if err != nil {
			g.logger.Error(err.Error())
		}
		return nil, err
	}

	return dps, nil
}

// NewGetAllExecutor creates a new GetProfilesExecutor for retrieving all device profiles.
func NewGetAllExecutor(
	config bootstrapConfig.ServiceInfo,
	loader DeviceProfileLoader,
	logger logger.LoggingClient) GetProfilesExecutor {

	return getAllDeviceProfiles{
		config: config,
		loader: loader,
		logger: logger,
	}
}

// getDeviceProfilesByModel encapsulates the data needed in order to get devices profiles by model.
type getDeviceProfilesByModel struct {
	model  string
	loader DeviceProfileLoader
}

// Execute retrieves device profiles by model.
func (g getDeviceProfilesByModel) Execute() ([]contract.DeviceProfile, error) {
	return g.loader.GetDeviceProfilesByModel(g.model)
}

// NewGetModelExecutor creates a new GetProfilesExecutor for retrieving device profiles by model.
func NewGetModelExecutor(model string, loader DeviceProfileLoader) GetProfilesExecutor {
	return getDeviceProfilesByModel{
		model:  model,
		loader: loader,
	}
}

// getDeviceProfilesWithLabel encapsulates the data needed in order to get devices profiles by label.
type getDeviceProfilesWithLabel struct {
	label  string
	loader DeviceProfileLoader
}

// Execute retrieves device profiles by label.
func (g getDeviceProfilesWithLabel) Execute() ([]contract.DeviceProfile, error) {
	return g.loader.GetDeviceProfilesWithLabel(g.label)
}

// NewGetLabelExecutor creates a new GetProfilesExecutor for retrieving device profiles by label.
func NewGetLabelExecutor(label string, loader DeviceProfileLoader) GetProfilesExecutor {
	return getDeviceProfilesWithLabel{
		label:  label,
		loader: loader,
	}
}

// getDeviceProfilesByManufacturerModel encapsulates the data needed in order to get device's profiles by
// manufacturer and model.
type getDeviceProfilesByManufacturerModel struct {
	manufacturer string
	model        string
	loader       DeviceProfileLoader
}

// Execute retrieves device profiles by manufacturer and model.
func (g getDeviceProfilesByManufacturerModel) Execute() ([]contract.DeviceProfile, error) {
	return g.loader.GetDeviceProfilesByManufacturerModel(g.manufacturer, g.model)
}

// NewGetManufacturerModelExecutor creates a new GetProfilesExecutor for retrieving device profiles by manufacturer
// and model.
func NewGetManufacturerModelExecutor(
	manufacturer string,
	model string,
	loader DeviceProfileLoader) GetProfilesExecutor {

	return getDeviceProfilesByManufacturerModel{
		manufacturer: manufacturer,
		model:        model,
		loader:       loader,
	}
}

// getDeviceProfilesByManufacturer encapsulates the data needed in order to get devices profiles by manufacturer.
type getDeviceProfilesByManufacturer struct {
	manufacturer string
	loader       DeviceProfileLoader
}

// Execute retrieves device profiles by manufacturer.
func (g getDeviceProfilesByManufacturer) Execute() ([]contract.DeviceProfile, error) {
	return g.loader.GetDeviceProfilesByManufacturer(g.manufacturer)
}

// NewGetManufacturerExecutor creates a new GetProfilesExecutor for retrieving device profiles by manufacturer.
func NewGetManufacturerExecutor(manufacturer string, loader DeviceProfileLoader) GetProfilesExecutor {
	return getDeviceProfilesByManufacturer{
		manufacturer: manufacturer,
		loader:       loader,
	}
}

// getProfileID encapsulates the data needed in order to get a devices profile by ID.
type getProfileID struct {
	id     string
	loader DeviceProfileLoader
}

// Execute retrieves a device profile by ID.
func (g getProfileID) Execute() (contract.DeviceProfile, error) {
	return g.loader.GetDeviceProfileById(g.id)
}

// NewGetProfileID creates a new GetProfileExecutor for retrieving device profiles by ID.
func NewGetProfileID(id string, loader DeviceProfileLoader) GetProfileExecutor {
	return getProfileID{
		id:     id,
		loader: loader,
	}
}

// getProfileName encapsulates the data needed in order to get a devices profile by name.
type getProfileName struct {
	name   string
	loader DeviceProfileLoader
}

// Execute retrieves a device profile by name.
func (g getProfileName) Execute() (contract.DeviceProfile, error) {
	return g.loader.GetDeviceProfileByName(g.name)
}

// NewGetProfileName creates a new GetProfileExecutor for retrieving device profiles by name.
func NewGetProfileName(name string, loader DeviceProfileLoader) GetProfileExecutor {
	return getProfileName{
		name:   name,
		loader: loader,
	}
}
