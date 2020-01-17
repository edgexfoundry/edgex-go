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
	"sync"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// UnimplementedMethodPanicMessage is the common panic message used for unimplemented methods of memory.Client.
const UnimplementedMethodPanicMessage = "This method has not been implemented."

// addressableMemoryStore encapsulates the data needed to store addressables.
type addressableMemoryStore struct {
	// addressableMapMutex provides a guard around the addressable store to avoid race conditions.
	// NOTE: it is up to the implementation to ensure locking and unlocking are being performed correctly.
	addressableMapMutex sync.RWMutex

	// addressableMap stores and organizes the persisted addressables.
	// Access to the this map should be done behind the addressableMapMutex with the proper locks being acquired before
	// accessing data.
	addressableMap map[string]contract.Addressable
}

// Client provides persistence using short-term storage for the underlying data-store.
//
// NOTE: Anything persisted will be lost if the managing process is shutdown or halted.
type Client struct {
	addressableStore addressableMemoryStore
}

// NewClient constructs a new Client for short-term storage.
func NewClient() *Client {
	return &Client{
		addressableStore: addressableMemoryStore{
			addressableMapMutex: sync.RWMutex{},
			addressableMap:      make(map[string]contract.Addressable),
		},
	}
}

// CloseSession cleans up resources in the underlying data-store.
//
// NOTE: Since this is a short-term data storage mechanism any persisted data is deleted.
func (c *Client) CloseSession() {
	c.addressableStore.addressableMapMutex.Lock()
	defer c.addressableStore.addressableMapMutex.Unlock()
	c.addressableStore.addressableMap = make(map[string]contract.Addressable)
}
