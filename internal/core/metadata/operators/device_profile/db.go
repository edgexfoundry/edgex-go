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

type DeviceLoader interface {
	GetAllDevices() ([]contract.Device, error)
	GetDevicesByProfileId(pid string) ([]contract.Device, error)
}

type DeviceProfileLoader interface {
	GetDeviceProfileById(id string) (contract.DeviceProfile, error)
	GetDeviceProfileByName(n string) (contract.DeviceProfile, error)
}

type DeviceProfileDeleter interface {
	DeleteDeviceProfileById(id string) error
	DeviceProfileLoader
	DeviceLoader
	ProvisionWatcherLoader
}

type DeviceProfileUpdater interface {
	GetAllDeviceProfiles() ([]contract.DeviceProfile, error)
	GetProvisionWatchersByProfileId(id string) ([]contract.ProvisionWatcher, error)
	UpdateDeviceProfile(dp contract.DeviceProfile) error
	DeviceProfileLoader
	DeviceLoader
}

type ProvisionWatcherLoader interface {
	GetProvisionWatchersByProfileId(id string) ([]contract.ProvisionWatcher, error)
}
