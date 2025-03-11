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
	"github.com/openziti/edge-api/rest_model"
	"github.com/sirupsen/logrus"
	"runtime"
)

var appId string
var appVersion string

func SetApplication(theAppId, theAppVersion string) {
	appId = theAppId
	appVersion = theAppVersion
}

func GetSdkInfo() (*rest_model.EnvInfo, *rest_model.SdkInfo) {
	sdkInfo := &rest_model.SdkInfo{
		AppID:      appId,
		AppVersion: appVersion,
		Type:       "ziti-sdk-golang",
		Version:    Version,
	}

	envInfo := &rest_model.EnvInfo{
		Arch:      runtime.GOARCH,
		Os:        runtime.GOOS,
		OsRelease: "",
		OsVersion: "",
	}

	if rel, ver, err := getOSversion(); err == nil {
		envInfo.OsRelease = rel
		envInfo.OsVersion = ver
	} else {
		logrus.Warn("failed to get OS version", err)
	}

	return envInfo, sdkInfo

}
