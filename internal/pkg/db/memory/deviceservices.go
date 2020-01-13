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

func (c *Client) GetDeviceServiceByName(n string) (contract.DeviceService, error) {
	return contract.DeviceService{}, nil
}

func (c *Client) GetDeviceServiceById(id string) (contract.DeviceService, error) {
	return contract.DeviceService{}, nil
}

func (c *Client) GetAllDeviceServices() ([]contract.DeviceService, error) {
	return []contract.DeviceService{}, nil
}

func (c *Client) GetDeviceServicesByAddressableId(id string) ([]contract.DeviceService, error) {
	return []contract.DeviceService{}, nil
}

func (c *Client) GetDeviceServicesWithLabel(l string) ([]contract.DeviceService, error) {
	return []contract.DeviceService{}, nil
}

func (c *Client) AddDeviceService(ds contract.DeviceService) (string, error) {
	return "", nil
}

func (c *Client) UpdateDeviceService(ds contract.DeviceService) error {
	return nil
}

func (c *Client) DeleteDeviceServiceById(id string) error {
	return nil
}
