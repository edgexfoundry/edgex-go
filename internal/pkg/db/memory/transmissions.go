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

func (c *Client) AddTransmission(t contract.Transmission) (string, error) {
	return "", nil
}

func (c *Client) UpdateTransmission(t contract.Transmission) error {
	return nil
}

func (c *Client) DeleteTransmission(age int64, status contract.TransmissionStatus) error {
	return nil
}

func (c *Client) GetTransmissionById(id string) (contract.Transmission, error) {
	return contract.Transmission{}, nil
}

func (c *Client) GetTransmissionsByNotificationSlug(slug string, limit int) ([]contract.Transmission, error) {
	return []contract.Transmission{}, nil
}

func (c *Client) GetTransmissionsByNotificationSlugAndStartEnd(slug string, start int64, end int64, limit int) ([]contract.Transmission, error) {
	return []contract.Transmission{}, nil
}

func (c *Client) GetTransmissionsByStartEnd(start int64, end int64, limit int) ([]contract.Transmission, error) {
	return []contract.Transmission{}, nil
}

func (c *Client) GetTransmissionsByStart(start int64, limit int) ([]contract.Transmission, error) {
	return []contract.Transmission{}, nil
}

func (c *Client) GetTransmissionsByEnd(end int64, limit int) ([]contract.Transmission, error) {
	return []contract.Transmission{}, nil
}

func (c *Client) GetTransmissionsByStatus(limit int, status contract.TransmissionStatus) ([]contract.Transmission, error) {
	return []contract.Transmission{}, nil
}
