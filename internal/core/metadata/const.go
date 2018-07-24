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
package metadata

// Struct used to pase the JSON configuration file
type ConfigurationStruct struct {
	DBType                              string
	MongoDatabaseName                   string
	MongoDBUserName                     string
	MongoDBPassword                     string
	MongoDBHost                         string
	MongoDBPort                         int
	MongoDBConnectTimeout               int
	ReadMaxLimit                        int
	Protocol                            string
	ServiceAddress                      string
	ServicePort                         int
	ServiceTimeout                      int
	AppOpenMsg                          string
	CheckInterval                       string
	ConsulProfilesActive                string
	ConsulHost                          string
	ConsulCheckAddress                  string
	ConsulPort                          int
	EnableRemoteLogging                 bool
	LoggingFile                         string
	LoggingRemoteURL                    string
	NotificationPostDeviceChanges       bool
	NotificationsSlug                   string
	NotificationContent                 string
	NotificationSender                  string
	NotificationDescription             string
	NotificationLabel                   string
	SupportNotificationsHost            string
	SupportNotificationsPort            int
	SupportNotificationsSubscriptionURL string
	SupportNotificationsTransmissionURL string
	StartupTimeout                      int
}

const (
	MAX_LIMIT int = 1000

	/* ---------------- URL PARAM NAMES -----------------------*/
	ID                       = "id"
	NAME                     = "name"
	OPSTATE                  = "opstate"
	URLADMINSTATE            = "adminstate"
	ADMINSTATE               = "adminState"
	URLLASTREPORTED          = "lastreported"
	LASTREPORTED             = "lastReported"
	LASTREPORTEDNOTIFY       = "lastreportednotify"
	URLLASTCONNECTED         = "lastconnected"
	LASTCONNECTED            = "lastConnected"
	LASTCONNECTEDNOTIFY      = "lastconnectednotify"
	ADDRESSABLE              = "addressable"
	ADDRESSABLENAME          = "addressablename"
	ADDRESSABLEID            = "addressableid"
	CHECK                    = "check"
	SERVICE                  = "service"
	SERVICENAME              = "servicename"
	SERVICEID                = "serviceid"
	LABEL                    = "label"
	PROFILE                  = "profile"
	PROFILEID                = "profileid"
	PROFILENAME              = "profilename"
	DEVICEPROFILE            = "deviceprofile"
	UPLOADFILE               = "uploadfile"
	UPLOAD                   = "upload"
	MODEL                    = "model"
	MANUFACTURER             = "manufacturer"
	YAML                     = "yaml"
	DEVICEREPORT             = "devicereport"
	DEVICENAME               = "devicename"
	DEVICESERVICE            = "deviceservice"
	SCHEDULEEVENT            = "scheduleevent"
	SCHEDULE                 = "schedule"
	TOPIC                    = "topic"
	PORT                     = "port"
	PUBLISHER                = "publisher"
	ADDRESS                  = "address"
	COMMAND                  = "command"
	DEVICE                   = "device"
	PROVISIONWATCHER         = "provisionwatcher"
	IDENTIFIER               = "identifier"
	KEY                      = "key"
	VALUE                    = "value"
	VALUEDESCRIPTORSFOR      = "valueDescriptorsFor"
	DEVICEADDRESSABLES       = "deviceaddressables"
	DEVICEADDRESSABLESBYNAME = "deviceaddressablesbyname"

	/* TODO ENUM */
	LOCKED   = "LOCKED"
	UNLOCKED = "UNLOCKED"
	ENABLED  = "ENABLED"
	DISABLED = "DISABLED"
	QUERYTS  = "-timestamp"
)
