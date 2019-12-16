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

package config

import (
	"fmt"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

func ListDefaultServices() map[string]string {
	return map[string]string{
		clients.SupportNotificationsServiceKey: "Notifications",
		clients.CoreCommandServiceKey:          "Command",
		clients.CoreDataServiceKey:             "CoreData",
		clients.CoreMetaDataServiceKey:         "Metadata",
		clients.ExportClientServiceKey:         "Export",
		clients.ExportDistroServiceKey:         "Distro",
		clients.SupportLoggingServiceKey:       "Logging",
		clients.SupportSchedulerServiceKey:     "Scheduler",
	}
}

// MessageQueueInfo provides parameters related to connecting to a message queue
type MessageQueueInfo struct {
	// Host is the hostname or IP address of the broker, if applicable.
	Host string
	// Port defines the port on which to access the message queue.
	Port int
	// Protocol indicates the protocol to use when accessing the message queue.
	Protocol string
	// Indicates the message queue platform being used.
	Type string
	// Indicates the topic the data is published/subscribed
	Topic string
}

func (m MessageQueueInfo) Uri() string {
	uri := fmt.Sprintf("%s://%s:%v", m.Protocol, m.Host, m.Port)
	return uri
}

// DatabaseInfo defines the parameters necessary for connecting to the desired persistence layer.
type DatabaseInfo map[string]bootstrapConfig.Database

type IntervalInfo struct {
	// Name of the schedule must be unique?
	Name string
	// Start time in ISO 8601 format YYYYMMDD'T'HHmmss
	Start string
	// End time in ISO 8601 format YYYYMMDD'T'HHmmss
	End string
	// Periodicity of the schedule
	Frequency string
	// Cron style regular expression indicating how often the action under schedule should occur.  Use either runOnce, frequency or cron and not all.
	Cron string
	// Boolean indicating that this schedules runs one time - at the time indicated by the start
	RunOnce bool
}

type IntervalActionInfo struct {
	// Host is the hostname or IP address of a service.
	Host string
	// Port defines the port on which to access a given service
	Port int
	// Protocol indicates the protocol to use when accessing a given service
	Protocol string
	// Action name
	Name string
	// Action http method *const prob*
	Method string
	// Acton target name
	Target string
	// Action target parameters
	Parameters string
	// Action target API path
	Path string
	// Associated Schedule for the Event
	Interval string
}

// ScheduleEventInfo helper function
func (e IntervalActionInfo) Url() string {
	url := fmt.Sprintf("%s://%s:%v", e.Protocol, e.Host, e.Port)
	return url
}

// Notification Info provides properties related to the assembly of notification content
type NotificationInfo struct {
	Content           string
	Description       string
	Label             string
	PostDeviceChanges bool
	Sender            string
	Slug              string
}
