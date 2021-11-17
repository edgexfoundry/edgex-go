//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package types

import "fmt"

const (
	defaultMsgProtocol = "tcp"
)

// HostInfo is the URL information of the host as the following scheme:
// <Protocol>://<Host>:<Port>
type HostInfo struct {
	// Host is the hostname or IP address of the messaging broker, if applicable.
	Host string
	// Port defines the port on which to access the message queue.
	Port int
	// Protocol indicates the protocol to use when accessing the message queue.
	Protocol string
}

// GetHostURL returns the complete URL for the host-info configuration
func (info *HostInfo) GetHostURL() string {

	protocol := info.Protocol
	if info.Protocol == "" {
		protocol = defaultMsgProtocol
	}
	return fmt.Sprintf("%s://%s:%d", protocol, info.Host, info.Port)
}

// IsHostInfoEmpty returns whether the host-info has been initialized or not
func (info *HostInfo) IsHostInfoEmpty() bool {
	return info == nil || info.Host == "" || info.Port == 0
}
