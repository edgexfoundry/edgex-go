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

package posture

import (
	"regexp"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

// OsProvider supplies operating system type and version information for OS posture checks.
type OsProvider interface {
	GetOsInfo() OsInfo
}

// OsInfo contains the operating system type and semantic version.
type OsInfo struct {
	Type    string
	Version string
}

// OsProviderFunc is a function adapter that implements OsProvider.
type OsProviderFunc func() OsInfo

func (f OsProviderFunc) GetOsInfo() OsInfo {
	return f()
}

// NewOsProvider creates the default OS information provider that queries system details.
func NewOsProvider() OsProvider {
	return &DefaultOsProvider{}
}

// DefaultOsProvider queries platform information to determine OS type and version.
type DefaultOsProvider struct{}

func (provider DefaultOsProvider) GetOsInfo() OsInfo {
	osType := runtime.GOOS
	osVersion := "unknown"

	semVerParser := regexp.MustCompile(`^((\d+)\.(\d+)\.(\d+))`)

	_, family, version, _ := host.PlatformInformation()

	if runtime.GOOS == "windows" {
		if strings.EqualFold(family, "server") {
			osType = "windowsserver"
		} else {
			osType = "windows"
		}
	}

	parsedVersion := semVerParser.FindStringSubmatch(version)

	if len(parsedVersion) > 1 {
		osVersion = parsedVersion[0]
	}
	return OsInfo{
		Type:    osType,
		Version: osVersion,
	}
}
