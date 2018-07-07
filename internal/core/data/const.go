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
 *******************************************************************************/
package data

type ConfigurationStruct struct {
	ConsulProfilesActive       string
	ReadMaxLimit               int
	MetaDataCheck              bool
	ValidateCheck              bool
	AddToEventQueue            bool
	PersistData                bool
	AppOpenMsg                 string
	FormatSpecifier            string
	MsgPubType                 string
	ServicePort                int
	ServiceTimeout             int
	ServiceAddress             string
	DeviceUpdateLastConnected  bool
	ServiceUpdateLastConnected bool
	DBType                     string
	MongoDBUserName            string
	MongoDBPassword            string
	MongoDatabaseName          string
	MongoDBHost                string
	MongoDBPort                int
	MongoDBConnectTimeout      int
	ConsulHost                 string
	ConsulCheckAddress         string
	ConsulPort                 int
	CheckInterval              string
	EnableRemoteLogging        bool
	LoggingFile                string
	LoggingRemoteURL           string
	MetaAddressableURL         string
	MetaAddressablePath        string
	MetaDeviceServiceURL       string
	MetaDeviceServicePath      string
	MetaDeviceProfileURL       string
	MetaDeviceProfilePath      string
	MetaDeviceURL              string
	MetaDevicePath             string
	MetaDeviceReportURL        string
	MetaDeviceReportPath       string
	MetaCommandURL             string
	MetaCommandPath            string
	MetaEventURL               string
	MetaEventPath              string
	MetaScheduleURL            string
	MetaSchedulePath           string
	MetaProvisionWatcherURL    string
	MetaProvisionWatcherPath   string
	MetaPingURL                string
	MetaPingPath               string
	ActiveMQBroker             string
	ZeroMQAddressPort          string
}

var configuration ConfigurationStruct = ConfigurationStruct{} //  Needs to be initialized before used
