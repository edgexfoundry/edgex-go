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

import (
	"github.com/google/uuid"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// GetAddressables retrieves all the stored addressables.
func (c *Client) GetAddressables() ([]contract.Addressable, error) {
	// Locking is handled in the getAllAddressables method as it is the one performing the read operation.
	return c.getAllAddressables(), nil
}

// UpdateAddressable updates the value stored for the provided addressable. If no addressable exists then an error is
// returned.
func (c *Client) UpdateAddressable(a contract.Addressable) error {
	c.addressableStore.addressableMapMutex.Lock()
	defer c.addressableStore.addressableMapMutex.Unlock()

	_, isPresent := c.addressableStore.addressableMap[a.Id]
	if !isPresent {
		return db.ErrNotFound
	}

	c.addressableStore.addressableMap[a.Id] = a
	return nil
}

// GetAddressableById retrieves the addressable with the associated ID. If no matching addressable exists then an error
// is returned.
func (c *Client) GetAddressableById(id string) (contract.Addressable, error) {
	c.addressableStore.addressableMapMutex.RLock()
	defer c.addressableStore.addressableMapMutex.RUnlock()

	addressable, isPresent := c.addressableStore.addressableMap[id]
	if !isPresent {
		return contract.Addressable{}, db.ErrNotFound
	}

	return addressable, nil
}

// AddAddressable stores the new addressable. If an addressable with a matching ID is already stored then an error is
// returned.
func (c *Client) AddAddressable(a contract.Addressable) (string, error) {
	c.addressableStore.addressableMapMutex.Lock()
	defer c.addressableStore.addressableMapMutex.Unlock()

	_, isPresent := c.addressableStore.addressableMap[a.Id]
	if isPresent {
		return "", db.ErrNotUnique
	}

	_, err := uuid.Parse(a.Id)
	if err != nil {
		a.Id = uuid.New().String()
	}

	c.addressableStore.addressableMap[a.Id] = a
	return a.Id, nil
}

// GetAddressableByName retrieves the addressables which match the provided name. If no matching addressable exists then
// an error is returned.
func (c *Client) GetAddressableByName(n string) (contract.Addressable, error) {
	c.addressableStore.addressableMapMutex.RLock()
	defer c.addressableStore.addressableMapMutex.RUnlock()

	for _, addressable := range c.addressableStore.addressableMap {
		if addressable.Name == n {
			return addressable, nil
		}
	}
	return contract.Addressable{}, db.ErrNotFound
}

// GetAddressablesByTopic retrieves the addressables which match the provided Topic. If no matching addressable exists
// then an error is returned.
func (c *Client) GetAddressablesByTopic(t string) ([]contract.Addressable, error) {
	c.addressableStore.addressableMapMutex.RLock()
	defer c.addressableStore.addressableMapMutex.RUnlock()

	matchingAddressables := make([]contract.Addressable, 0)
	for _, addressable := range c.addressableStore.addressableMap {
		if addressable.Topic == t {
			matchingAddressables = append(matchingAddressables, addressable)
		}
	}

	if len(matchingAddressables) > 0 {
		return matchingAddressables, nil
	}

	return matchingAddressables, db.ErrNotFound
}

// GetAddressablesByPort retrieves the addressables which match the provided Port. If no matching addressable exists
// then an error is returned.
func (c *Client) GetAddressablesByPort(p int) ([]contract.Addressable, error) {
	c.addressableStore.addressableMapMutex.RLock()
	defer c.addressableStore.addressableMapMutex.RUnlock()

	matchingAddressables := make([]contract.Addressable, 0)
	for _, addressable := range c.addressableStore.addressableMap {
		if addressable.Port == p {
			matchingAddressables = append(matchingAddressables, addressable)
		}
	}

	if len(matchingAddressables) > 0 {
		return matchingAddressables, nil
	}

	return matchingAddressables, db.ErrNotFound
}

// GetAddressablesByPublisher retrieves the addressables which match the provided Publisher. If no matching addressable
// exists then an error is returned.
func (c *Client) GetAddressablesByPublisher(p string) ([]contract.Addressable, error) {
	c.addressableStore.addressableMapMutex.RLock()
	defer c.addressableStore.addressableMapMutex.RUnlock()

	matchingAddressables := make([]contract.Addressable, 0)
	for _, addressable := range c.addressableStore.addressableMap {
		if addressable.Publisher == p {
			matchingAddressables = append(matchingAddressables, addressable)
		}
	}

	if len(matchingAddressables) > 0 {
		return matchingAddressables, nil
	}

	return matchingAddressables, db.ErrNotFound
}

// GetAddressablesByAddress retrieves the addressables which match the provided Address. If no matching addressable
// exists then an error is returned.
func (c *Client) GetAddressablesByAddress(add string) ([]contract.Addressable, error) {
	c.addressableStore.addressableMapMutex.RLock()
	defer c.addressableStore.addressableMapMutex.RUnlock()

	matchingAddressables := make([]contract.Addressable, 0)
	for _, addressable := range c.addressableStore.addressableMap {
		if addressable.Address == add {
			matchingAddressables = append(matchingAddressables, addressable)
		}
	}

	if len(matchingAddressables) > 0 {
		return matchingAddressables, nil
	}

	return matchingAddressables, db.ErrNotFound
}

// DeleteAddressableById removes the addressable with the associated ID from memory. If no matching addressable exists
// then an error is returned.
func (c *Client) DeleteAddressableById(id string) error {
	c.addressableStore.addressableMapMutex.Lock()
	defer c.addressableStore.addressableMapMutex.Unlock()

	_, isPresent := c.addressableStore.addressableMap[id]
	if !isPresent {
		return db.ErrNotFound
	}

	delete(c.addressableStore.addressableMap, id)
	return nil
}

// getAllAddressables retrieves all the persisted addressables and returns them in a slice.
//
// NOTE: Acquires a read lock when invoked.
func (c *Client) getAllAddressables() []contract.Addressable {
	c.addressableStore.addressableMapMutex.RLock()
	defer c.addressableStore.addressableMapMutex.RUnlock()

	addressables := make([]contract.Addressable, 0)
	for _, value := range c.addressableStore.addressableMap {
		addressables = append(addressables, value)
	}

	return addressables
}
