/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package clients

const ClientMonitorDefault = 15000

const (
	ApiAddressableRoute        = "/api/v1/addressable"
	ApiCommandRoute            = "/api/v1/command"
	ApiDeviceRoute             = "/api/v1/device"
	ApiDeviceProfileRoute      = "/api/v1/deviceprofile"
	ApiDeviceServiceRoute      = "/api/v1/deviceservice"
	ApiEventRoute              = "/api/v1/event"
	ApiLoggingRoute            = "/api/v1/logs"
	ApiNotificationRoute       = "/api/v1/notification"
	ApiNotifyRegistrationRoute = "/api/v1/notify/registrations"
	ApiProvisionWatcherRoute   = "/api/v1/provisionwatcher"
	ApiRegistrationRoute       = "/api/v1/registration"
	ApiRegistrationByNameRoute = ApiRegistrationRoute + "/name"
	ApiScheduleRoute           = "/api/v1/schedule"
	ApiScheduleEventRoute      = "/api/v1/scheduleevent"
	ApiSubscriptionRoute       = "/api/v1/subscription"
	ApiTransmissionRoute       = "/api/v1/transmission"
	ApiConfigRoute             = "/api/v1/config"
	ApiMetricsRoute            = "/api/v1/metrics"
	ApiPingRoute               = "/api/v1/ping"
)
