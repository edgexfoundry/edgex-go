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
	Applicationname            string
	Consulprofilesactive       string
	Readmaxlimit               int
	Metadatacheck              bool
	Addtoeventqueue            bool
	Persistdata                bool
	HeartBeatTime              int
	HeartBeatMsg               string
	Appopenmsg                 string
	Formatspecifier            string
	Msgpubtype                 string
	Serverport                 int
	Serviceaddress             string
	Servicename                string
	Deviceupdatelastconnected  bool
	Serviceupdatelastconnected bool
	Datamongodbusername        string
	Datamongodbpassword        string
	Datamongodbdatabase        string
	Datamongodbhost            string
	Datamongodbport            int
	DatamongodbsocketTimeout   int
	DatamongodbmaxWaitTime     int
	DatamongodbsocketKeepAlive bool
	Consulhost                 string
	Consulcheckaddress         string
	Consulport                 int
	Checkinterval              string
	EnableRemoteLogging        bool
	Loggingfile                string
	Loggingremoteurl           string
	Metadbaddressableurl       string
	Metadbdeviceserviceurl     string
	Metadbdeviceprofileurl     string
	Metadbdeviceurl            string
	Metadbdevicereporturl      string
	Metadbcommandurl           string
	Metadbeventurl             string
	Metadbscheduleurl          string
	Metadbprovisionwatcherurl  string
	Metadbpingurl              string
	Activemqbroker             string
	Zeromqaddressport          string
	Amqbroker                  string
}

var configuration ConfigurationStruct = ConfigurationStruct{} //  Needs to be initialized before used

var (
	COREDATASERVICENAME = "core-data"
)