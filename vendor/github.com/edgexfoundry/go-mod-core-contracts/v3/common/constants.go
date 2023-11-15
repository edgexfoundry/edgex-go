//
// Copyright (C) 2020-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

// Constants related to defined routes in the v3 service APIs
const (
	ApiVersion = "v3"
	ApiBase    = "/api/v3"

	ApiEventRoute                                           = ApiBase + "/event"
	ApiEventServiceNameProfileNameDeviceNameSourceNameRoute = ApiEventRoute + "/{" + ServiceName + "}" + "/{" + ProfileName + "}" + "/{" + DeviceName + "}" + "/{" + SourceName + "}"
	ApiAllEventRoute                                        = ApiEventRoute + "/" + All
	ApiEventIdRoute                                         = ApiEventRoute + "/" + Id + "/{" + Id + "}"
	ApiEventCountRoute                                      = ApiEventRoute + "/" + Count
	ApiEventCountByDeviceNameRoute                          = ApiEventCountRoute + "/" + Device + "/" + Name + "/{" + Name + "}"
	ApiEventByDeviceNameRoute                               = ApiEventRoute + "/" + Device + "/" + Name + "/{" + Name + "}"
	ApiEventByTimeRangeRoute                                = ApiEventRoute + "/" + Start + "/{" + Start + "}/" + End + "/{" + End + "}"
	ApiEventByAgeRoute                                      = ApiEventRoute + "/" + Age + "/{" + Age + "}"

	ApiReadingRoute                                        = ApiBase + "/reading"
	ApiAllReadingRoute                                     = ApiReadingRoute + "/" + All
	ApiReadingCountRoute                                   = ApiReadingRoute + "/" + Count
	ApiReadingCountByDeviceNameRoute                       = ApiReadingCountRoute + "/" + Device + "/" + Name + "/{" + Name + "}"
	ApiReadingByDeviceNameRoute                            = ApiReadingRoute + "/" + Device + "/" + Name + "/{" + Name + "}"
	ApiReadingByResourceNameRoute                          = ApiReadingRoute + "/" + ResourceName + "/{" + ResourceName + "}"
	ApiReadingByTimeRangeRoute                             = ApiReadingRoute + "/" + Start + "/{" + Start + "}/" + End + "/{" + End + "}"
	ApiReadingByResourceNameAndTimeRangeRoute              = ApiReadingByResourceNameRoute + "/" + Start + "/{" + Start + "}/" + End + "/{" + End + "}"
	ApiReadingByDeviceNameAndResourceNameRoute             = ApiReadingRoute + "/" + Device + "/" + Name + "/{" + Name + "}/" + ResourceName + "/{" + ResourceName + "}"
	ApiReadingByDeviceNameAndResourceNameAndTimeRangeRoute = ApiReadingByDeviceNameAndResourceNameRoute + "/" + Start + "/{" + Start + "}/" + End + "/{" + End + "}"
	ApiReadingByDeviceNameAndTimeRangeRoute                = ApiReadingByDeviceNameRoute + "/" + Start + "/{" + Start + "}/" + End + "/{" + End + "}"

	ApiDeviceProfileRoute                       = ApiBase + "/deviceprofile"
	ApiDeviceProfileBasicInfoRoute              = ApiDeviceProfileRoute + "/basicinfo"
	ApiDeviceProfileDeviceCommandRoute          = ApiDeviceProfileRoute + "/" + DeviceCommand
	ApiDeviceProfileResourceRoute               = ApiDeviceProfileRoute + "/" + Resource
	ApiDeviceProfileUploadFileRoute             = ApiDeviceProfileRoute + "/uploadfile"
	ApiDeviceProfileByNameRoute                 = ApiDeviceProfileRoute + "/" + Name + "/{" + Name + "}"
	ApiDeviceProfileDeviceCommandByNameRoute    = ApiDeviceProfileByNameRoute + "/" + DeviceCommand + "/{" + CommandName + "}"
	ApiDeviceProfileResourceByNameRoute         = ApiDeviceProfileByNameRoute + "/" + Resource + "/{" + ResourceName + "}"
	ApiDeviceProfileByIdRoute                   = ApiDeviceProfileRoute + "/" + Id + "/{" + Id + "}"
	ApiAllDeviceProfileRoute                    = ApiDeviceProfileRoute + "/" + All
	ApiDeviceProfileByManufacturerRoute         = ApiDeviceProfileRoute + "/" + Manufacturer + "/{" + Manufacturer + "}"
	ApiDeviceProfileByModelRoute                = ApiDeviceProfileRoute + "/" + Model + "/{" + Model + "}"
	ApiDeviceProfileByManufacturerAndModelRoute = ApiDeviceProfileRoute + "/" + Manufacturer + "/{" + Manufacturer + "}" + "/" + Model + "/{" + Model + "}"

	ApiDeviceResourceRoute                     = ApiBase + "/deviceresource"
	ApiDeviceResourceByProfileAndResourceRoute = ApiDeviceResourceRoute + "/" + Profile + "/{" + ProfileName + "}" + "/" + Resource + "/{" + ResourceName + "}"

	ApiDeviceServiceRoute       = ApiBase + "/deviceservice"
	ApiAllDeviceServiceRoute    = ApiDeviceServiceRoute + "/" + All
	ApiDeviceServiceByNameRoute = ApiDeviceServiceRoute + "/" + Name + "/{" + Name + "}"
	ApiDeviceServiceByIdRoute   = ApiDeviceServiceRoute + "/" + Id + "/{" + Id + "}"

	ApiDeviceRoute                = ApiBase + "/device"
	ApiAllDeviceRoute             = ApiDeviceRoute + "/" + All
	ApiDeviceIdExistsRoute        = ApiDeviceRoute + "/" + Check + "/" + Id + "/{" + Id + "}"
	ApiDeviceNameExistsRoute      = ApiDeviceRoute + "/" + Check + "/" + Name + "/{" + Name + "}"
	ApiDeviceByIdRoute            = ApiDeviceRoute + "/" + Id + "/{" + Id + "}"
	ApiDeviceByNameRoute          = ApiDeviceRoute + "/" + Name + "/{" + Name + "}"
	ApiDeviceByProfileIdRoute     = ApiDeviceRoute + "/" + Profile + "/" + Id + "/{" + Id + "}"
	ApiDeviceByProfileNameRoute   = ApiDeviceRoute + "/" + Profile + "/" + Name + "/{" + Name + "}"
	ApiDeviceByServiceIdRoute     = ApiDeviceRoute + "/" + Service + "/" + Id + "/{" + Id + "}"
	ApiDeviceByServiceNameRoute   = ApiDeviceRoute + "/" + Service + "/" + Name + "/{" + Name + "}"
	ApiDeviceNameCommandNameRoute = ApiDeviceByNameRoute + "/{" + Command + "}"

	ApiProvisionWatcherRoute              = ApiBase + "/provisionwatcher"
	ApiAllProvisionWatcherRoute           = ApiProvisionWatcherRoute + "/" + All
	ApiProvisionWatcherByIdRoute          = ApiProvisionWatcherRoute + "/" + Id + "/{" + Id + "}"
	ApiProvisionWatcherByNameRoute        = ApiProvisionWatcherRoute + "/" + Name + "/{" + Name + "}"
	ApiProvisionWatcherByProfileNameRoute = ApiProvisionWatcherRoute + "/" + Profile + "/" + Name + "/{" + Name + "}"
	ApiProvisionWatcherByServiceNameRoute = ApiProvisionWatcherRoute + "/" + Service + "/" + Name + "/{" + Name + "}"

	ApiSubscriptionRoute           = ApiBase + "/subscription"
	ApiAllSubscriptionRoute        = ApiSubscriptionRoute + "/" + All
	ApiSubscriptionByNameRoute     = ApiSubscriptionRoute + "/" + Name + "/{" + Name + "}"
	ApiSubscriptionByCategoryRoute = ApiSubscriptionRoute + "/" + Category + "/{" + Category + "}"
	ApiSubscriptionByLabelRoute    = ApiSubscriptionRoute + "/" + Label + "/{" + Label + "}"
	ApiSubscriptionByReceiverRoute = ApiSubscriptionRoute + "/" + Receiver + "/{" + Receiver + "}"

	ApiNotificationCleanupRoute            = ApiBase + "/cleanup"
	ApiNotificationCleanupByAgeRoute       = ApiBase + "/" + Cleanup + "/" + Age + "/{" + Age + "}"
	ApiNotificationRoute                   = ApiBase + "/notification"
	ApiNotificationByTimeRangeRoute        = ApiNotificationRoute + "/" + Start + "/{" + Start + "}/" + End + "/{" + End + "}"
	ApiNotificationByAgeRoute              = ApiNotificationRoute + "/" + Age + "/{" + Age + "}"
	ApiNotificationByCategoryRoute         = ApiNotificationRoute + "/" + Category + "/{" + Category + "}"
	ApiNotificationByLabelRoute            = ApiNotificationRoute + "/" + Label + "/{" + Label + "}"
	ApiNotificationByIdRoute               = ApiNotificationRoute + "/" + Id + "/{" + Id + "}"
	ApiNotificationByStatusRoute           = ApiNotificationRoute + "/" + Status + "/{" + Status + "}"
	ApiNotificationBySubscriptionNameRoute = ApiNotificationRoute + "/" + Subscription + "/" + Name + "/{" + Name + "}"

	ApiTransmissionRoute                   = ApiBase + "/transmission"
	ApiTransmissionByIdRoute               = ApiTransmissionRoute + "/" + Id + "/{" + Id + "}"
	ApiTransmissionByAgeRoute              = ApiTransmissionRoute + "/" + Age + "/{" + Age + "}"
	ApiAllTransmissionRoute                = ApiTransmissionRoute + "/" + All
	ApiTransmissionBySubscriptionNameRoute = ApiTransmissionRoute + "/" + Subscription + "/" + Name + "/{" + Name + "}"
	ApiTransmissionByTimeRangeRoute        = ApiTransmissionRoute + "/" + Start + "/{" + Start + "}/" + End + "/{" + End + "}"
	ApiTransmissionByStatusRoute           = ApiTransmissionRoute + "/" + Status + "/{" + Status + "}"
	ApiTransmissionByNotificationIdRoute   = ApiTransmissionRoute + "/" + Notification + "/" + Id + "/{" + Id + "}"

	ApiConfigRoute         = ApiBase + "/config"
	ApiPingRoute           = ApiBase + "/ping"
	ApiVersionRoute        = ApiBase + "/version"
	ApiSecretRoute         = ApiBase + "/secret"
	ApiUnitsOfMeasureRoute = ApiBase + "/uom"

	ApiDeviceCallbackRoute      = ApiBase + "/callback/device"
	ApiDeviceCallbackNameRoute  = ApiBase + "/callback/device/name/{name}"
	ApiProfileCallbackRoute     = ApiBase + "/callback/profile"
	ApiProfileCallbackNameRoute = ApiBase + "/callback/profile/name/{name}"
	ApiWatcherCallbackRoute     = ApiBase + "/callback/watcher"
	ApiWatcherCallbackNameRoute = ApiBase + "/callback/watcher/name/{name}"
	ApiServiceCallbackRoute     = ApiBase + "/callback/service"
	ApiDiscoveryRoute           = ApiBase + "/discovery"
	ApiDeviceValidationRoute    = ApiBase + "/validate/device"

	ApiIntervalRoute               = ApiBase + "/interval"
	ApiAllIntervalRoute            = ApiIntervalRoute + "/" + All
	ApiIntervalByNameRoute         = ApiIntervalRoute + "/" + Name + "/{" + Name + "}"
	ApiIntervalActionRoute         = ApiBase + "/intervalaction"
	ApiAllIntervalActionRoute      = ApiIntervalActionRoute + "/" + All
	ApiIntervalActionByNameRoute   = ApiIntervalActionRoute + "/" + Name + "/{" + Name + "}"
	ApiIntervalActionByTargetRoute = ApiIntervalActionRoute + "/" + Target + "/{" + Target + "}"

	ApiSystemRoute      = ApiBase + "/system"
	ApiOperationRoute   = ApiSystemRoute + "/operation"
	ApiHealthRoute      = ApiSystemRoute + "/health"
	ApiMultiConfigRoute = ApiSystemRoute + "/config"
)

// Constants related to defined url path names and parameters in the v3 service APIs
const (
	All           = "all"
	Id            = "id"
	Created       = "created"
	Modified      = "modified"
	Pushed        = "pushed"
	Origin        = "origin"
	Count         = "count"
	Device        = "device"
	DeviceId      = "deviceId"
	DeviceName    = "deviceName"
	DeviceCommand = "deviceCommand"
	Check         = "check"
	Profile       = "profile"
	Resource      = "resource"
	Service       = "service"
	Services      = "services"
	Command       = "command"
	ProfileName   = "profileName"
	SourceName    = "sourceName"
	ServiceName   = "serviceName"
	ResourceName  = "resourceName"
	ResourceNames = "resourceNames"
	CommandName   = "commandName"
	Start         = "start"
	End           = "end"
	Age           = "age"
	Scrub         = "scrub"
	Type          = "type"
	Name          = "name"
	Label         = "label"
	Manufacturer  = "manufacturer"
	Model         = "model"
	ValueType     = "valueType"
	Category      = "category"
	Receiver      = "receiver"
	Subscription  = "subscription"
	Notification  = "notification"
	Target        = "target"
	Status        = "status"
	Cleanup       = "cleanup"
	Sender        = "sender"
	Severity      = "severity"
	Interval      = "interval"

	Offset       = "offset"         //query string to specify the number of items to skip before starting to collect the result set.
	Limit        = "limit"          //query string to specify the numbers of items to return
	Labels       = "labels"         //query string to specify associated user-defined labels for querying a given object. More than one label may be specified via a comma-delimited list
	PushEvent    = "ds-pushevent"   //query string to specify if an event should be pushed to the EdgeX system
	ReturnEvent  = "ds-returnevent" //query string to specify if an event should be returned from device service
	RegexCommand = "ds-regexcmd"    //query string to specify if the command name is in regular expression format
)

// Constants related to the default value of query strings in the v3 service APIs
const (
	DefaultOffset  = 0
	DefaultLimit   = 20
	CommaSeparator = ","
	ValueTrue      = "true"
	ValueFalse     = "false"
)

// Constants related to Reading ValueTypes
const (
	ValueTypeBool         = "Bool"
	ValueTypeString       = "String"
	ValueTypeUint8        = "Uint8"
	ValueTypeUint16       = "Uint16"
	ValueTypeUint32       = "Uint32"
	ValueTypeUint64       = "Uint64"
	ValueTypeInt8         = "Int8"
	ValueTypeInt16        = "Int16"
	ValueTypeInt32        = "Int32"
	ValueTypeInt64        = "Int64"
	ValueTypeFloat32      = "Float32"
	ValueTypeFloat64      = "Float64"
	ValueTypeBinary       = "Binary"
	ValueTypeBoolArray    = "BoolArray"
	ValueTypeStringArray  = "StringArray"
	ValueTypeUint8Array   = "Uint8Array"
	ValueTypeUint16Array  = "Uint16Array"
	ValueTypeUint32Array  = "Uint32Array"
	ValueTypeUint64Array  = "Uint64Array"
	ValueTypeInt8Array    = "Int8Array"
	ValueTypeInt16Array   = "Int16Array"
	ValueTypeInt32Array   = "Int32Array"
	ValueTypeInt64Array   = "Int64Array"
	ValueTypeFloat32Array = "Float32Array"
	ValueTypeFloat64Array = "Float64Array"
	ValueTypeObject       = "Object"
)

// Constants related to configuration file's map key
const (
	Primary  = "Primary"
	Password = "Password"
)

// Constants for Address
const (
	// Type
	REST  = "REST"
	MQTT  = "MQTT"
	EMAIL = "EMAIL"
)

// Constants for SMA Operation Action
const (
	ActionStart   = "start"
	ActionRestart = "restart"
	ActionStop    = "stop"
)

// Constants for DeviceProfile
const (
	ReadWrite_R  = "R"
	ReadWrite_W  = "W"
	ReadWrite_RW = "RW"
	ReadWrite_WR = "WR"
)

// Constants for Edgex Environment variable
const (
	EnvEncodeAllEvents = "EDGEX_ENCODE_ALL_EVENTS_CBOR"
)

// Miscellaneous constants
const (
	ClientMonitorDefault = 15000              // Defaults the interval at which a given service client will refresh its endpoint from the Registry, if used
	CorrelationHeader    = "X-Correlation-ID" // Sets the key of the Correlation ID HTTP header
)

// Constants related to how services identify themselves in the Service Registry
const (
	CoreCommandServiceKey               = "core-command"
	CoreDataServiceKey                  = "core-data"
	CoreMetaDataServiceKey              = "core-metadata"
	CoreCommonConfigServiceKey          = "core-common-config-bootstrapper"
	SupportLoggingServiceKey            = "support-logging"
	SupportNotificationsServiceKey      = "support-notifications"
	SystemManagementAgentServiceKey     = "sys-mgmt-agent"
	SupportSchedulerServiceKey          = "support-scheduler"
	SecuritySecretStoreSetupServiceKey  = "security-secretstore-setup"
	SecurityProxyAuthServiceKey         = "security-proxy-auth"
	SecurityProxySetupServiceKey        = "security-proxy-setup"
	SecurityFileTokenProviderServiceKey = "security-file-token-provider"
	SecurityBootstrapperKey             = "security-bootstrapper"
	SecurityBootstrapperRedisKey        = "security-bootstrapper-redis"
	SecuritySpiffeTokenProviderKey      = "security-spiffe-token-provider" // nolint:gosec
)

// Constants related to the possible content types supported by the APIs
const (
	Accept          = "Accept"
	ContentType     = "Content-Type"
	ContentLength   = "Content-Length"
	ContentTypeCBOR = "application/cbor"
	ContentTypeJSON = "application/json"
	ContentTypeTOML = "application/toml"
	ContentTypeYAML = "application/x-yaml"
	ContentTypeText = "text/plain"
	ContentTypeXML  = "application/xml"
)

// Constants related to System Events
const (
	DeviceSystemEventType           = "device"
	DeviceProfileSystemEventType    = "deviceprofile"
	DeviceServiceSystemEventType    = "deviceservice"
	ProvisionWatcherSystemEventType = "provisionwatcher"
	SystemEventActionAdd            = "add"
	SystemEventActionUpdate         = "update"
	SystemEventActionDelete         = "delete"
)

const (
	ConfigStemAll      = "edgex/v3" // Version never changes during minor releases so v3 is more appropriate than 3.0
	ConfigStemApp      = ConfigStemAll
	ConfigStemCore     = ConfigStemAll
	ConfigStemDevice   = ConfigStemAll
	ConfigStemSecurity = ConfigStemAll
)

const (
	CommandQueryRequestTopicKey   = "CommandQueryRequestTopic" // #nosec G101
	CommandQueryResponseTopicKey  = "CommandQueryResponseTopic"
	CommandRequestTopicKey        = "CommandRequestTopic"
	CommandResponseTopicPrefixKey = "CommandResponseTopicPrefix"
)

// MessageBus Topics
const (

	// Common Topics
	DefaultBaseTopic        = "edgex"         // Used if the base topic is not specified in MessageBus configuration
	EventsPublishTopic      = "events"        // <ServiceType>/<DeviceServiceName>/<ProfileName>/<DeviceName>/<SourceName> are appended
	ResponseTopic           = "response"      // <ServiceName>/<RequestId> are prepended
	MetricsPublishTopic     = "telemetry"     // <ServiceName>/<MetricName> are prepended
	SystemEventPublishTopic = "system-events" // <SourceServiceName>/<SystemEventType>/<SystemEventAction><OwnerServiceName>/<ProfileName>

	// Core Data Topics
	CoreDataEventSubscribeTopic = "events/device/#"

	// Core Command Topics
	CoreCommandDeviceRequestPublishTopic  = "device/command/request" // <DeviceServiceName>/<DeviceName>/<CommandName>/<CommandMethod> are appended
	CoreCommandRequestSubscribeTopic      = "core/command/request/#"
	CoreCommandQueryRequestSubscribeTopic = "core/commandquery/request/#"

	// Command Client Topics
	CoreCommandQueryRequestPublishTopic = "core/commandquery/request" // <deviceName>|all is prepended
	CoreCommandRequestPublishTopic      = "core/command/request"      // <DeviceName>/<CommandName>/<CommandMethod> are appended

	// Support Notifications
	// No Topics Yet

	// Support Scheduler
	// No Topics Yet

	// Device Services Topics
	CommandRequestSubscribeTopic      = "device/command/request"          // <DeviceServiceName>/# is appended <
	MetadataSystemEventSubscribeTopic = "system-events/core-metadata/+/+" // <DeviceServiceName>/# is appended
	ValidateDeviceSubscribeTopic      = "validate/device"                 // <DeviceServiceName> is pre-pended

	// App Service Topics
	// App Service topics remain configurable inorder to filter by subscription.
)
