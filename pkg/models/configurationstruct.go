//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package models

type ConfigurationStruct struct {
	Hostname             string
	Port                 int
	DistroHost           string
	ClientHost           string
	DataHost             string
	DataPort             int
	ConsulHost           string
	ConsulPort           int
	ConsulProfilesActive string
	CheckInterval        string
	MQTTSCert            string
	MQTTSKey             string
	MarkPushed           bool
}
