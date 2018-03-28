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
 * @microservice: core-metadata-go service
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package metadata

import (
	"errors"

	"github.com/edgexfoundry/edgex-go/core/domain/enums"
)

// Struct used to pase the JSON configuration file
type ConfigurationStruct struct {
	ApplicationName                     string
	DBType                              string
	MongoDatabaseName                   string
	MongoDBUserName                     string
	MongoDBPassword                     string
	MongoDBHost                         string
	MongoDBPort                         int
	MongoDBConnectTimeout               int
	ReadMaxLimit                        int
	Protocol                            string
	ServiceName                         string
	ServiceAddress                      string
	ServicePort                          int
	ServiceTimeout                       int
	HeartBeatTime                       int
	HeartBeatMsg                        string
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
	SupportNotificationsNotificationURL string
	SupportNotificationsSubscriptionURL string
	SupportNotificationsTransmissionURL string
}

// Configuration data for the metadata service
var configuration ConfigurationStruct = ConfigurationStruct{} // Needs to be initialized before used

var (
	/* -------------- CONFIG for METADATA -------------------- */
	DATABASE            enums.DATABASE
	DBTYPE              = "mongodb"
	PROTOCOL            = "http"
	SERVERPORT          = "48081"
	DOCKERMONGO         = "edgex-mongo:27017"
	DBUSER              = "meta"
	DBPASS              = "password"
	MONGODATABASE       = "metadata"
	METADATASERVICENAME = "core-metadata"

	MAX_LIMIT int = 1000

	/* ----------------------- CONSTANTS ----------------------------*/
	REST       = "http"
	MONGOSTR   = "mongo"
	DB         = "metadata"
	DEVICECOL  = "device"
	DPCOL      = "deviceProfile"
	DSCOL      = "deviceService"
	ADDCOL     = "addressable"
	COMCOL     = "command"
	DRCOL      = "deviceReport"
	SECOL      = "scheduleEvent"
	SCOL       = "schedule"
	PWCOL      = "provisionWatcher"
	TIMELAYOUT = "20060102T150405"

	/* ---------------- URL PARAM NAMES -----------------------*/
	ID                       = "id"
	_ID                      = "_id"
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
	SERVICE                  = "service"
	SERVICENAME              = "servicename"
	SERVICEID                = "serviceid"
	LABEL                    = "label"
	LABELS                   = "labels"
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
	OPERATINGSTATE           = "operatingState"
	PROVISIONWATCHER         = "provisionwatcher"
	IDENTIFIER               = "identifier"
	IDENTIFIERS              = "identifiers"
	KEY                      = "key"
	VALUE                    = "value"
	VALUEDESCRIPTORSFOR      = "valueDescriptorsFor"
	DEVICEADDRESSABLES       = "deviceaddressables"
	DEVICEADDRESSABLESBYNAME = "deviceaddressablesbyname"

	/* ----------------------- ERRORS ----------------------------*/
	ErrNotFound                  = errors.New("Not found")
	ErrDuplicateName             = errors.New("Duplicate name for the resource")
	ErrDuplicateCommandInProfile = errors.New("Duplicate name for command in device profile")
	ErrCommandStillInUse         = errors.New("Command is still in use by device profiles")
	/* TODO ENUM */
	LOCKED   = "LOCKED"
	UNLOCKED = "UNLOCKED"
	ENABLED  = "ENABLED"
	DISABLED = "DISABLED"
	QUERYTS  = "-timestamp"
)
