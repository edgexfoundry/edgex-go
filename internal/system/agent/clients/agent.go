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

	"github.com/edgexfoundry/go-mod-core-contracts/clients/agent"
)

// AgentType is an alias for agent.AgentClient and hides implementation detail.
type AgentType agent.AgentClient

// clientMap defines internal map structure to track multiple instances of agent.AgentClient.
type clientMap map[string]AgentType

// Agent contains implementation structures for tracking multiple instances of AgentType.
type Agent struct {
	clients clientMap
	mutex   sync.RWMutex
}

// NewAgent is a factory function that returns an initialized Agent receiver struct.
func NewAgent() *Agent {
	return &Agent{
		clients: make(clientMap),
		mutex:   sync.RWMutex{},
	}
}

// Get returns the AgentType and ok = true for the requested client name if it exists, otherwise ok = false.
func (c *Agent) Get(clientName string) (client AgentType, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	client, ok = c.clients[clientName]
	return
}

// Set updates the list of clients to ensure the provided clientName key contains the provided AgentType value.
func (c *Agent) Set(clientName string, value AgentType) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clients[clientName] = value
}
