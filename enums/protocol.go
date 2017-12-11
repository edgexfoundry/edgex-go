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

package enums

/*
This file is the protocol enum for EdgeX
Current values: HTTP, TCP, MAC, ZMQ, OTHER, SSL
HTTP - for REST communications
TCP - for MQTT and other general TCP based communications
MAC - MAC address - low level (example serial) communications
ZMQ - Zero MQ communications
SSL - for TLS encrypted sockets
 */

type ProtocolType uint8

const(
	HTTP ProtocolType = iota
	TCP
    MAC
    ZMQ
    OTHER
    SSL
)

var protocolStringArray = [...]string{"HTTP", "TCP", "MAC", "ZMQ", "OTHER", "SSL"}

/*
String() function for formatting
 */
func (p ProtocolType) String() string{
	return protocolStringArray[p]
}
