/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package agent

import (
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
)

// clientMap defines internal map structure to track multiple instances of general.GeneralClient.
type clientMap map[string]general.GeneralClient

// GeneralClients contains implementation structures for tracking multiple instances of general.GeneralClient.
type GeneralClients struct {
	clients clientMap
	mutex   sync.RWMutex
}

// NewGeneralClients is a factory function that returns an initialized GeneralClients receiver struct.
func NewGeneralClients() *GeneralClients {
	return &GeneralClients{
		clients: make(clientMap),
		mutex:   sync.RWMutex{},
	}
}

// Get returns the general.GeneralClient and ok = true for the requested client name if it exists, otherwise ok = false.
func (c *GeneralClients) Get(clientName string) (client general.GeneralClient, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	client, ok = c.clients[clientName]
	return
}

// Set updates the list of clients to ensure the provided clientName key contains the provided general.GeneralClient value.
func (c *GeneralClients) Set(clientName string, value general.GeneralClient) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clients[clientName] = value
}
