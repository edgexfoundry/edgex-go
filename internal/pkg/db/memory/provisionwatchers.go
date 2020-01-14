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

func (c *Client) GetAllProvisionWatchers() (pw []contract.ProvisionWatcher, err error) {
	return []contract.ProvisionWatcher{}, nil
}

func (c *Client) GetProvisionWatcherByName(n string) (pw contract.ProvisionWatcher, err error) {
	return contract.ProvisionWatcher{}, nil
}

func (c *Client) GetProvisionWatchersByIdentifier(k string, v string) (pw []contract.ProvisionWatcher, err error) {
	return []contract.ProvisionWatcher{}, nil
}

func (c *Client) GetProvisionWatchersByServiceId(id string) (pw []contract.ProvisionWatcher, err error) {
	return []contract.ProvisionWatcher{}, nil
}

func (c *Client) GetProvisionWatchersByProfileId(id string) (pw []contract.ProvisionWatcher, err error) {
	return []contract.ProvisionWatcher{}, nil
}

func (c *Client) GetProvisionWatcherById(id string) (pw contract.ProvisionWatcher, err error) {
	return contract.ProvisionWatcher{}, nil
}

func (c *Client) AddProvisionWatcher(pw contract.ProvisionWatcher) (string, error) {
	return "", nil
}

func (c *Client) UpdateProvisionWatcher(pw contract.ProvisionWatcher) error {
	return nil
}

func (c *Client) DeleteProvisionWatcherById(id string) error {
	return nil
}
