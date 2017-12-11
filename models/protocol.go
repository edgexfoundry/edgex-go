/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *
 * @microservice: core-domain-go library
 * @author: Ryan Comer & Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/

package models

import (
	"encoding/json"
	"fmt"
)

type Protocol byte

/*
 * HTTP - for REST communications
 * TCP - for MQTT and other general TCP based communications
 * MAC - MAC address - low level (example serial) communications
 * ZMQ - Zero MQ communications
 * SSL - for TLS encrypted sockets
 */
const (
	HTTP Protocol = iota
	TCP
	MAC
	ZMQ
	OTHER
    SSL
)
/*
 * Unmarshaller for enum type
 */
func (p *Protocol) UnmarshalJSON(data []byte) error {
	// Extract the string from data.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Protocol should be a string, got %s", data)
	}

	got, err := map[string]Protocol{"HTTP": HTTP, "TCP": TCP, "MAC": MAC, "ZMQ": ZMQ, "OTHER": OTHER, "SSL": SSL}[s]
	if !err {
		return fmt.Errorf("invalid Protocol %q", s)
	}
	*p = got
	return nil
}