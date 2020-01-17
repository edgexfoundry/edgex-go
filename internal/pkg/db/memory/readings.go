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

func (c *Client) Readings() ([]contract.Reading, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddReading(r contract.Reading) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateReading(r contract.Reading) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ReadingById(id string) (contract.Reading, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ReadingCount() (int, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteReadingById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteReadingsByDevice(deviceId string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ReadingsByDevice(id string, limit int) ([]contract.Reading, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ReadingsByValueDescriptor(name string, limit int) ([]contract.Reading, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ReadingsByValueDescriptorNames(names []string, limit int) ([]contract.Reading, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ReadingsByCreationTime(start, end int64, limit int) ([]contract.Reading, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]contract.Reading, error) {
	panic(UnimplementedMethodPanicMessage)
}
