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

package clients

import (
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
)

// GeneralType is an alias for general.GeneralClient and hides implementation detail.
type GeneralType general.GeneralClient

// clientMap defines internal map structure to track multiple instances of general.GeneralClient.
type clientMap map[string]GeneralType

// General contains implementation structures for tracking multiple instances of GeneralType.
type General struct {
	clients clientMap
	mutex   sync.RWMutex
}

// NewGeneral is a factory function that returns an initialized General receiver struct.
func NewGeneral() *General {
	return &General{
		clients: make(clientMap),
		mutex:   sync.RWMutex{},
	}
}

// Get returns the GeneralType and ok = true for the requested client name if it exists, otherwise ok = false.
func (c *General) Get(clientName string) (client GeneralType, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	client, ok = c.clients[clientName]
	return
}

// Set updates the list of clients to ensure the provided clientName key contains the provided GeneralType value.
func (c *General) Set(clientName string, value GeneralType) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clients[clientName] = value
}
