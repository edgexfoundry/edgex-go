/*
	Copyright NetFoundry Inc.

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

package identity

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func parseAddr(addr string) (*url.URL, error) {
	if isFile(addr) {
		addr = strings.TrimPrefix(addr, "file:")
		return &url.URL{
			Scheme: "file",
			Path:   strings.TrimPrefix(addr, "//"),
		}, nil
	}

	if isPem(addr) {
		addr = strings.TrimPrefix(addr, "pem:")
		return &url.URL{
			Scheme: "pem",
			Opaque: strings.TrimPrefix(addr, "//"),
		}, nil
	}

	if isEngine(addr) {
		u := strings.SplitN(addr, ":", 2)
		if len(u) > 1 {
			pathAndArgs := strings.SplitN(u[1], "?", 2)
			return &url.URL{
				Scheme:   u[0],
				Host:     strings.TrimPrefix(pathAndArgs[0], "//"),
				RawQuery: pathAndArgs[1],
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to parse address [%s]", addr)
}

// addr is determined to be a file if:
// - begins with "file:"
// - satisfies filepath.IsAbs()
// - does not contain a ":" (at least we know addr is not a "pem:" or "<engine>:" )
func isFile(addr string) bool {
	return strings.HasPrefix(addr, "file:") || filepath.IsAbs(addr) || !strings.Contains(addr, ":")
}

func isPem(addr string) bool {
	return strings.HasPrefix(addr, "pem:")
}

func isEngine(addr string) bool {
	return strings.Contains(addr, ":") && !isFile(addr)
}
