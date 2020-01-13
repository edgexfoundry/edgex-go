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

func (c *Client) GetAddressables() ([]contract.Addressable, error) {
	return []contract.Addressable{}, nil
}

func (c *Client) UpdateAddressable(a contract.Addressable) error {
	return nil
}

func (c *Client) GetAddressableById(id string) (contract.Addressable, error) {
	return contract.Addressable{}, nil
}

func (c *Client) AddAddressable(a contract.Addressable) (string, error) {
	return "", nil
}

func (c *Client) GetAddressableByName(n string) (contract.Addressable, error) {
	return contract.Addressable{}, nil
}

func (c *Client) GetAddressablesByTopic(t string) ([]contract.Addressable, error) {
	return []contract.Addressable{}, nil
}

func (c *Client) GetAddressablesByPort(p int) ([]contract.Addressable, error) {
	return []contract.Addressable{}, nil
}

func (c *Client) GetAddressablesByPublisher(p string) ([]contract.Addressable, error) {
	return []contract.Addressable{}, nil
}

func (c *Client) GetAddressablesByAddress(add string) ([]contract.Addressable, error) {
	return []contract.Addressable{}, nil
}

func (c *Client) DeleteAddressableById(id string) error {
	return nil
}
