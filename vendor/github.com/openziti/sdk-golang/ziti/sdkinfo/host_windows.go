/*
	Copyright 2020 NetFoundry Inc.

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

package sdkinfo

import (
	"fmt"
	"syscall"
)

func getOSversion() (string, string, error) {
	if ver, err := syscall.GetVersion(); err == nil {
		major := ver & 0xff
		minor := (ver >> 8) & 0xff
		buildnum := (ver >> 16)

		rel := fmt.Sprintf("%d.%d.%d", major, minor, buildnum)
		return rel, rel, nil
	} else {
		return "", "", err
	}
}
