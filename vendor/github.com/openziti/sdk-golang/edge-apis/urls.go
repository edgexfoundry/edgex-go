/*
	Copyright 2019 NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package edge_apis

import (
	"strings"
)

const (
	ClientApiPath     = "/edge/client/v1"
	ManagementApiPath = "/edge/management/v1"
)

// ClientUrl returns a URL with the given hostname in the format of `https://<hostname>/edge/management/v1`.
// The hostname provided may include a port.
func ClientUrl(hostname string) string {
	return concat(hostname, ClientApiPath)
}

// ManagementUrl returns a URL with the given hostname in the format of `https://<hostname>/edge/management/v1`.
// The hostname provided may include a port.
func ManagementUrl(hostname string) string {
	return concat(hostname, ManagementApiPath)
}

func concat(base, path string) string {
	if !strings.Contains(base, "://") {
		base = "https://" + base
	}
	if strings.HasSuffix(base, "/") {
		return strings.Trim(base, "/") + path
	}
	return base + path
}
