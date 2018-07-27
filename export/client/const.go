//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package client

type ConfigurationStruct struct {
	Hostname             string
	Port                 int
	DBType               string
	DBURL             string
	DBUsername        string
	DBPassword        string
	DBDatabase        string
	DBPort            int
	DBConnectTimeout  int
	DBSocketTimeout   int
	ConsulHost           string
	ConsulPort           int
	CheckInterval        string
	ConsulProfilesActive string
	DistroHost           string
	DistroPort           int
}

var configuration ConfigurationStruct = ConfigurationStruct{} // Needs to be initialized before used

