//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

type ConfigurationStruct struct {
	ApplicationName      string
	Hostname             string
	Port                 int
	DistroHost           string
	ClientHost           string
	DataHost             string
	ConsulHost           string
	ConsulPort           int
	ConsulProfilesActive string
	CheckInterval        string
	MQTTSCert            string
	MQTTSKey             string
}

var configuration ConfigurationStruct = ConfigurationStruct{} // Needs to be initialized before used
