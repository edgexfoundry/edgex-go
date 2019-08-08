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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// DeviceLoader retrieves devices as needed from the perspective of device profiles.
type DeviceLoader interface {
	GetAllDevices() ([]contract.Device, error)
	GetDevicesByProfileId(pid string) ([]contract.Device, error)
}

// DeviceProfileAdder adds DeviceProfiles to the database
type DeviceProfileAdder interface {
	AddDeviceProfile(d contract.DeviceProfile) (string, error)
}

// DeviceProfileLoader retrieves device profiles.
type DeviceProfileLoader interface {
	GetDeviceProfileById(id string) (contract.DeviceProfile, error)
	GetDeviceProfileByName(n string) (contract.DeviceProfile, error)
	GetAllDeviceProfiles() ([]contract.DeviceProfile, error)
	GetDeviceProfilesByModel(m string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesWithLabel(l string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesByManufacturerModel(man string, mod string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesByManufacturer(man string) ([]contract.DeviceProfile, error)
}

// DeviceProfileDeleter deletes device profiles.
// Also provides other functionality for validating the device profile before deletion. Such as loading other entities
// to ensure there are no other dependencies on the device profile before deletion.
type DeviceProfileDeleter interface {
	DeleteDeviceProfileById(id string) error

	// Functionality needed to perform validation and check state of DeviceProfile
	DeviceProfileLoader
	DeviceLoader
	ProvisionWatcherLoader
}

// DeviceProfileUpdater updates device profiles.
// Also provides other functionality for validating the device profile before deletion. Such as loading other entities
// to ensure there are no other dependencies on the device profile before deletion.
type DeviceProfileUpdater interface {
	GetProvisionWatchersByProfileId(id string) ([]contract.ProvisionWatcher, error)
	UpdateDeviceProfile(dp contract.DeviceProfile) error
	DeviceProfileLoader
	DeviceLoader
}

// ProvisionWatcherLoader retrieves provision watchers.
type ProvisionWatcherLoader interface {
	GetProvisionWatchersByProfileId(id string) ([]contract.ProvisionWatcher, error)
}
