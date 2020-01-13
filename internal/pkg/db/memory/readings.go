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
	return []contract.Reading{}, nil
}

func (c *Client) AddReading(r contract.Reading) (string, error) {
	return "", nil
}

func (c *Client) UpdateReading(r contract.Reading) error {
	return nil
}

func (c *Client) ReadingById(id string) (contract.Reading, error) {
	return contract.Reading{}, nil
}

func (c *Client) ReadingCount() (int, error) {
	return 0, nil
}

func (c *Client) DeleteReadingById(id string) error {
	return nil
}

func (c *Client) DeleteReadingsByDevice(deviceId string) error {
	return nil
}

func (c *Client) ReadingsByDevice(id string, limit int) ([]contract.Reading, error) {
	return []contract.Reading{}, nil
}

func (c *Client) ReadingsByValueDescriptor(name string, limit int) ([]contract.Reading, error) {
	return []contract.Reading{}, nil
}

func (c *Client) ReadingsByValueDescriptorNames(names []string, limit int) ([]contract.Reading, error) {
	return []contract.Reading{}, nil
}

func (c *Client) ReadingsByCreationTime(start, end int64, limit int) ([]contract.Reading, error) {
	return []contract.Reading{}, nil
}

func (c *Client) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]contract.Reading, error) {
	return []contract.Reading{}, nil
}
