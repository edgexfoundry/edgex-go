//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

// Constants related to defined routes with path params in the v3 service APIs for Echo
// TODO: Remove in EdgeX 4.0 and will use the original API constant names defined in constants.go
const (
	ApiEventServiceNameProfileNameDeviceNameSourceNameEchoRoute = ApiEventRoute + "/:" + ServiceName + "/:" + ProfileName + "/:" + DeviceName + "/:" + SourceName
	ApiEventIdEchoRoute                                         = ApiEventRoute + "/" + Id + "/:" + Id
	ApiEventCountByDeviceNameEchoRoute                          = ApiEventCountRoute + "/" + Device + "/" + Name + "/:" + Name
	ApiEventByDeviceNameEchoRoute                               = ApiEventRoute + "/" + Device + "/" + Name + "/:" + Name
	ApiEventByTimeRangeEchoRoute                                = ApiEventRoute + "/" + Start + "/:" + Start + "/" + End + "/:" + End
	ApiEventByAgeEchoRoute                                      = ApiEventRoute + "/" + Age + "/:" + Age

	ApiReadingCountByDeviceNameEchoRoute                       = ApiReadingCountRoute + "/" + Device + "/" + Name + "/:" + Name
	ApiReadingByDeviceNameEchoRoute                            = ApiReadingRoute + "/" + Device + "/" + Name + "/:" + Name
	ApiReadingByResourceNameEchoRoute                          = ApiReadingRoute + "/" + ResourceName + "/:" + ResourceName
	ApiReadingByTimeRangeEchoRoute                             = ApiReadingRoute + "/" + Start + "/:" + Start + "/" + End + "/:" + End
	ApiReadingByResourceNameAndTimeRangeEchoRoute              = ApiReadingByResourceNameEchoRoute + "/" + Start + "/:" + Start + "/" + End + "/:" + End
	ApiReadingByDeviceNameAndResourceNameEchoRoute             = ApiReadingRoute + "/" + Device + "/" + Name + "/:" + Name + "/" + ResourceName + "/:" + ResourceName
	ApiReadingByDeviceNameAndResourceNameAndTimeRangeEchoRoute = ApiReadingByDeviceNameAndResourceNameEchoRoute + "/" + Start + "/:" + Start + "/" + End + "/:" + End
	ApiReadingByDeviceNameAndTimeRangeEchoRoute                = ApiReadingByDeviceNameEchoRoute + "/" + Start + "/:" + Start + "/" + End + "/:" + End

	ApiDeviceProfileByNameEchoRoute                 = ApiDeviceProfileRoute + "/" + Name + "/:" + Name
	ApiDeviceProfileDeviceCommandByNameEchoRoute    = ApiDeviceProfileByNameEchoRoute + "/" + DeviceCommand + "/:" + CommandName
	ApiDeviceProfileResourceByNameEchoRoute         = ApiDeviceProfileByNameEchoRoute + "/" + Resource + "/:" + ResourceName
	ApiDeviceProfileByIdEchoRoute                   = ApiDeviceProfileRoute + "/" + Id + "/:" + Id
	ApiDeviceProfileByManufacturerEchoRoute         = ApiDeviceProfileRoute + "/" + Manufacturer + "/:" + Manufacturer
	ApiDeviceProfileByModelEchoRoute                = ApiDeviceProfileRoute + "/" + Model + "/:" + Model
	ApiDeviceProfileByManufacturerAndModelEchoRoute = ApiDeviceProfileRoute + "/" + Manufacturer + "/:" + Manufacturer + "/" + Model + "/:" + Model

	ApiDeviceResourceByProfileAndResourceEchoRoute = ApiDeviceResourceRoute + "/" + Profile + "/:" + ProfileName + "/" + Resource + "/:" + ResourceName

	ApiDeviceServiceByNameEchoRoute = ApiDeviceServiceRoute + "/" + Name + "/:" + Name
	ApiDeviceServiceByIdEchoRoute   = ApiDeviceServiceRoute + "/" + Id + "/:" + Id

	ApiDeviceIdExistsEchoRoute        = ApiDeviceRoute + "/" + Check + "/" + Id + "/:" + Id
	ApiDeviceNameExistsEchoRoute      = ApiDeviceRoute + "/" + Check + "/" + Name + "/:" + Name
	ApiDeviceByIdEchoRoute            = ApiDeviceRoute + "/" + Id + "/:" + Id
	ApiDeviceByNameEchoRoute          = ApiDeviceRoute + "/" + Name + "/:" + Name
	ApiDeviceByProfileIdEchoRoute     = ApiDeviceRoute + "/" + Profile + "/" + Id + "/:" + Id
	ApiDeviceByProfileNameEchoRoute   = ApiDeviceRoute + "/" + Profile + "/" + Name + "/:" + Name
	ApiDeviceByServiceIdEchoRoute     = ApiDeviceRoute + "/" + Service + "/" + Id + "/:" + Id
	ApiDeviceByServiceNameEchoRoute   = ApiDeviceRoute + "/" + Service + "/" + Name + "/:" + Name
	ApiDeviceNameCommandNameEchoRoute = ApiDeviceByNameEchoRoute + "/:" + Command

	ApiProvisionWatcherByIdEchoRoute          = ApiProvisionWatcherRoute + "/" + Id + "/:" + Id
	ApiProvisionWatcherByNameEchoRoute        = ApiProvisionWatcherRoute + "/" + Name + "/:" + Name
	ApiProvisionWatcherByProfileNameEchoRoute = ApiProvisionWatcherRoute + "/" + Profile + "/" + Name + "/:" + Name
	ApiProvisionWatcherByServiceNameEchoRoute = ApiProvisionWatcherRoute + "/" + Service + "/" + Name + "/:" + Name

	ApiSubscriptionByNameEchoRoute     = ApiSubscriptionRoute + "/" + Name + "/:" + Name
	ApiSubscriptionByCategoryEchoRoute = ApiSubscriptionRoute + "/" + Category + "/:" + Category
	ApiSubscriptionByLabelEchoRoute    = ApiSubscriptionRoute + "/" + Label + "/:" + Label
	ApiSubscriptionByReceiverEchoRoute = ApiSubscriptionRoute + "/" + Receiver + "/:" + Receiver

	ApiNotificationCleanupByAgeEchoRoute       = ApiBase + "/" + Cleanup + "/" + Age + "/:" + Age
	ApiNotificationByTimeRangeEchoRoute        = ApiNotificationRoute + "/" + Start + "/:" + Start + "/" + End + "/:" + End
	ApiNotificationByAgeEchoRoute              = ApiNotificationRoute + "/" + Age + "/:" + Age
	ApiNotificationByCategoryEchoRoute         = ApiNotificationRoute + "/" + Category + "/:" + Category
	ApiNotificationByLabelEchoRoute            = ApiNotificationRoute + "/" + Label + "/:" + Label
	ApiNotificationByIdEchoRoute               = ApiNotificationRoute + "/" + Id + "/:" + Id
	ApiNotificationByStatusEchoRoute           = ApiNotificationRoute + "/" + Status + "/:" + Status
	ApiNotificationBySubscriptionNameEchoRoute = ApiNotificationRoute + "/" + Subscription + "/" + Name + "/:" + Name

	ApiTransmissionByIdEchoRoute               = ApiTransmissionRoute + "/" + Id + "/:" + Id
	ApiTransmissionByAgeEchoRoute              = ApiTransmissionRoute + "/" + Age + "/:" + Age
	ApiTransmissionBySubscriptionNameEchoRoute = ApiTransmissionRoute + "/" + Subscription + "/" + Name + "/:" + Name
	ApiTransmissionByTimeRangeEchoRoute        = ApiTransmissionRoute + "/" + Start + "/:" + Start + "/" + End + "/:" + End
	ApiTransmissionByStatusEchoRoute           = ApiTransmissionRoute + "/" + Status + "/:" + Status
	ApiTransmissionByNotificationIdEchoRoute   = ApiTransmissionRoute + "/" + Notification + "/" + Id + "/:" + Id

	ApiDeviceCallbackNameEchoRoute  = ApiBase + "/callback/device/name/:name"
	ApiProfileCallbackNameEchoRoute = ApiBase + "/callback/profile/name/:name"
	ApiWatcherCallbackNameEchoRoute = ApiBase + "/callback/watcher/name/:name"

	ApiIntervalByNameEchoRoute         = ApiIntervalRoute + "/" + Name + "/:" + Name
	ApiIntervalActionByNameEchoRoute   = ApiIntervalActionRoute + "/" + Name + "/:" + Name
	ApiIntervalActionByTargetEchoRoute = ApiIntervalActionRoute + "/" + Target + "/:" + Target
)
