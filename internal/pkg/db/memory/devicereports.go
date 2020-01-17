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

func (c *Client) GetAllDeviceReports() ([]contract.DeviceReport, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceReportByName(n string) (contract.DeviceReport, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceReportByDeviceName(n string) ([]contract.DeviceReport, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceReportById(id string) (contract.DeviceReport, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetDeviceReportsByAction(n string) ([]contract.DeviceReport, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddDeviceReport(d contract.DeviceReport) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateDeviceReport(dr contract.DeviceReport) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteDeviceReportById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}
