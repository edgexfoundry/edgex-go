/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package memory

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

func (c *Client) GetAllDeviceProfiles() ([]contract.DeviceProfile, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceProfileById(id string) (contract.DeviceProfile, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceProfilesByModel(model string) ([]contract.DeviceProfile, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceProfilesWithLabel(l string) ([]contract.DeviceProfile, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceProfilesByManufacturerModel(man string, mod string) ([]contract.DeviceProfile, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceProfilesByManufacturer(man string) ([]contract.DeviceProfile, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceProfileByName(n string) (contract.DeviceProfile, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddDeviceProfile(dp contract.DeviceProfile) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateDeviceProfile(dp contract.DeviceProfile) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteDeviceProfileById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}
