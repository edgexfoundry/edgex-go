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
 * @microservice: core-command-go service
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package command

// ConfigurationStruct : Struct used to pase the JSON configuration file
type ConfigurationStruct struct {
	ApplicationName           string
	ConsulProfilesActive      string
	ReadMaxLimit              int
	ServicePort               int
	HeartBeatTime             int
	ConsulPort                int
	ServiceTimeout            int
	CheckInterval             string
	ServiceAddress            string
	ServiceName               string
	DeviceServiceProtocol     string
	HeartBeatMsg              string
	AppOpenMsg                string
	URLProtocol               string
	URLDevicePath             string
	ConsulHost                string
	ConsulCheckAddress        string
	EnableRemoteLogging       bool
	LogFile                   string
	LoggingRemoteURL          string
	MetaAddressableURL        string
	MetaDeviceServiceURL      string
	MetaDeviceProfileURL      string
	MetaDeviceURL             string
	MetaDeviceReportURL       string
	MetaCommandURL            string
	MetaEventURL              string
	MetaScheduleURL           string
	MetaProvisionWatcherURL   string
}

// Configuration data for the metadata service
var configuration ConfigurationStruct = ConfigurationStruct{}

const (
	/* -------------- Constants for Command -------------------- */
	COMMANDSERVICENAME       string = "core-command"
	REST_HTTP                string = "http://"
	ID                       string = "id"
	_ID                      string = "_id"
	NAME                     string = "name"
	DEVICEIDURLPARAM         string = "{deviceId}"
	OPSTATE                  string = "opstate"
	URLADMINSTATE            string = "adminstate"
	ADMINSTATE               string = "adminState"
	URLLASTREPORTED          string = "lastreported"
	LASTREPORTED             string = "lastReported"
	LASTREPORTEDNOTIFY       string = "lastreportednotify"
	URLLASTCONNECTED         string = "lastconnected"
	LASTCONNECTED            string = "lastConnected"
	LASTCONNECTEDNOTIFY      string = "lastconnectednotify"
	DEVICEMANAGER            string = "devicemanager"
	ADDRESSABLE              string = "addressable"
	ADDRESSABLENAME          string = "addressablename"
	ADDRESSABLEID            string = "addressableid"
	SERVICE                  string = "service"
	SERVICENAME              string = "servicename"
	SERVICEID                string = "serviceid"
	LABEL                    string = "label"
	LABELS                   string = "labels"
	PROFILE                  string = "profile"
	PROFILEID                string = "profileid"
	PROFILENAME              string = "profilename"
	DEVICEPROFILE            string = "deviceprofile"
	UPLOADFILE               string = "uploadfile"
	UPLOAD                   string = "upload"
	MODEL                    string = "model"
	MANUFACTURER             string = "manufacturer"
	YAML                     string = "yaml"
	DEVICEREPORT             string = "devicereport"
	DEVICENAME               string = "devicename"
	DEVICESERVICE            string = "deviceservice"
	SCHEDULEEVENT            string = "scheduleevent"
	SCHEDULE                 string = "schedule"
	TOPIC                    string = "topic"
	PORT                     string = "port"
	PUBLISHER                string = "publisher"
	ADDRESS                  string = "address"
	COMMAND                  string = "command"
	COMMANDID                string = "commandid"
	DEVICE                   string = "device"
	OPERATINGSTATE           string = "operatingState"
	PROVISIONWATCHER         string = "provisionwatcher"
	IDENTIFIER               string = "identifier"
	IDENTIFIERS              string = "identifiers"
	KEY                      string = "key"
	VALUE                    string = "value"
	VALUEDESCRIPTORSFOR             = "valueDescriptorsFor"
	DEVICEADDRESSABLES              = "deviceaddressables"
	DEVICEADDRESSABLESBYNAME        = "deviceaddressablesbyname"
	PINGENDPOINT                    = "/ping"
	PINGRESPONSE                    = "pong"
	CONTENTTYPE                     = "Content-Type"
	TEXTPLAIN                       = "text/plain"

	/* TODO ENUM */
	LOCKED   string = "LOCKED"
	UNLOCKED string = "UNLOCKED"
	ENABLED  string = "ENABLED"
	DISABLED string = "DISABLED"
	QUERYTS  string = "-timestamp"
)
