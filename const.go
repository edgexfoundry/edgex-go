package main

import "errors"

// Struct used to pase the JSON configuration file
type ConfigurationStruct struct {
	ApplicationName                     string
	MongoDBName                         string
	MongoDBUserName                     string
	MongoDBPassword                     string
	MongoDBHost                         string
	MongoDBPort                         int
	MongoDBConnectTimeout               int
	Readmaxlimit                        int
	Protocol                            string
	ServiceName                         string
	ServiceAddress                      string
	ServerPort                          int
	ServerTimeout                       int
	HeartBeatTime                       int
	HeartBeatMsg                        string
	AppOpenMsg                          string
	CheckInterval                       string
	ConsulProfilesActive                string
	Consulhost                          string
	Consulcheckaddress                  string
	ConsulPort                          int
	LoggingFile                         string
	LoggingRemoteURL                    string
	Notificationpostdevicechanges       bool
	Notificationslug                    string
	Notificationcontent                 string
	Notificationsender                  string
	Notificationdescription             string
	Notificationlabel                   string
	SupportNotificationsNotificationURL string
	SupportNotificationsSubscriptionURL string
	SupportNotificationsTransmissionURL string
}

// Configuration data for the metadata service
var configuration ConfigurationStruct = ConfigurationStruct{} // Needs to be initialized before used

var (
	/* -------------- CONFIG for METADATA -------------------- */
	DATABASE    = "mongo"
	PROTOCOL    = "http"
	SERVERPORT  = "48081"
	DOCKERMONGO = "fuse-mongo:27017"
	DBUSER      = "meta"
	DBPASS      = "password"

	MAX_LIMIT int = 1000

	/* ----------------------- CONSTANTS ----------------------------*/
	REST       = "http"
	GET        = "GET"
	PUT        = "PUT"
	POST       = "POST"
	DELETE     = "DELETE"
	MONGOSTR   = "mongo"
	DB         = "metadata"
	DEVICECOL  = "device"
	DPCOL      = "deviceProfile"
	DSCOL      = "deviceService"
	DMCOL      = "deviceManager"
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
	DEVICEMANAGER            = "devicemanager"
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
	LOCKED   = "locked"
	UNLOCKED = "unlocked"
	ENABLED  = "enabled"
	DISABLED = "disabled"
	QUERYTS  = "-timestamp"
)
