/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package data

type ConfigurationStruct struct {
	ApplicationName            string
	ConsulProfilesActive       string
	ReadMaxLimit               int
	MetaDataCheck              bool
	ValidateCheck              bool
	AddToEventQueue            bool
	PersistData                bool
	HeartBeatTime              int
	HeartBeatMsg               string
	AppOpenMsg                 string
	FormatSpecifier            string
	MsgPubType                 string
	ServicePort                int
	ServiceTimeout             int
	ServiceAddress             string
	ServiceName                string
	DeviceUpdateLastConnected  bool
	ServiceUpdateLastConnected bool
	MongoDBUserName            string
	MongoDBPassword            string
	MongoDatabaseName          string
	MongoDBHost                string
	MongoDBPort                int
	MongoDBConnectTimeout      int
	MongoDBMaxWaitTime         int
	MongoDBKeepAlive           bool
	ConsulHost                 string
	ConsulCheckAddress         string
	ConsulPort                 int
	CheckInterval              string
	EnableRemoteLogging        bool
	LoggingFile                string
	LoggingRemoteURL           string
	MetaAddressableURL         string
	MetaDeviceServiceURL       string
	MetaDeviceProfileURL       string
	MetaDeviceURL              string
	MetaDeviceReportURL        string
	MetaCommandURL             string
	MetaEventURL               string
	MetaScheduleURL            string
	MetaProvisionWatcherURL    string
	MetaPingURL                string
	ActiveMQBroker             string
	ZeroMQAddressPort          string
	AmqBroker                  string
}

var configuration ConfigurationStruct = ConfigurationStruct{} //  Needs to be initialized before used

var (
	COREDATASERVICENAME = "core-data"
)
