//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package client

type ConfigurationStruct struct {
	Hostname             string
	Port                 int
	Timeout              int
	DBType               string
	MongoURL             string
	MongoUsername        string
	MongoPassword        string
	MongoDatabaseName    string
	MongoPort            int
	MongoConnectTimeout  int
	MongoSocketTimeout   int
	ConsulHost           string
	ConsulPort           int
	CheckInterval        string
	ConsulProfilesActive string
	DistroHost           string
	DistroPort           int
	EnableRemoteLogging  bool
	LoggingRemoteURL     string
	LogFile              string
	AppOpenMsg           string
}

var configuration = ConfigurationStruct{} // Needs to be initialized before used
