//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

type ConfigurationStruct struct {
	ApplicationName      string
	Hostname             string
	Port                 int
	Persistence          string
	LogFilename          string
	MongoDB              string
	MongoCollection      string
	MongoURL             string
	MongoPort            int
	MongoConnectTimeout  int
	SocketTimeout        int
	MongoUsername        string
	MongoPassword        string
	CheckInterval        string
	ConsulHost           string
	ConsulPort           int
	ConsulProfilesActive string
}

// Configuration data for the support logging service
var configuration ConfigurationStruct = ConfigurationStruct{}
