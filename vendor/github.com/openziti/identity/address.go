//go:build !windows
// +build !windows

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
	"net/url"
	"strings"
)

// Parse identity certificate and key values, which are represented as URLs
// For example:
// - "file:///path/to/key.pem"
// - "pem:-----BEGIN CERTIFICATE-----\nMIICEjC..."
// - "engine:{engine_id}?{engine_opts}"
func parseAddr(addr string) (*url.URL, error) {
	// Don't try to parse "pem:" addresses which might contain newlines.
	// As of go 1.12, url.Parse returns an error when given URLs that contain control characters.
	// https://golang.org/doc/go1.12#net/url
	if strings.HasPrefix(addr, "pem:") || strings.HasPrefix(addr, "-----BEGIN") {
		ret := &url.URL{
			Scheme: "pem",
			Opaque: strings.TrimPrefix(addr, "pem:"),
		}
		return ret, nil
	}
	urlParts, err := url.Parse(addr)

	if err != nil {
		return nil, err
	}

	if urlParts.Scheme == "" {
		urlParts.Scheme = StorageFile
	}

	return urlParts, nil
}
