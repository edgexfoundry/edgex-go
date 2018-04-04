//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

type ConfigurationStruct struct {
	ApplicationName               string
	ReadLimit                     int
	ServerPort                    int
	ServerTimeout                 int
	HeartbeatTime                 int
	HeartbeatMsg                  string
	AppOpenMsg                    string
	ServiceName                   string
	ServiceHost                   string
	ServicePort                   int
	ServiceLabels                 string
	ServiceCallback               string
	ServiceConnectRetries         int
	ServiceConnectInterval        int
	ScheduleInterval              int
	ConsulHost                    string
	ConsulPort                    int
	CheckInterval                 string
	DefaultScheduleName           string
	DefaultScheduleFrequency      string
	DefaultScheduleStart          string
	DefaultScheduleEventName      string
	DefaultScheduleEventMethod    string
	DefaultScheduleEventService   string
	DefaultScheduleEventPath      string
	DefaultScheduleEventSchedule  string
	DefaultScheduleEventScheduler string
	EnableRemoteLogging           bool
	LoggingFile                   string
	LoggingRemoteUrl              string
	Metadbaddressableurl          string
	Metadbdeviceserviceurl        string
	Metadbdeviceprofileurl        string
	Metadbdeviceurl               string
	Metadbdevicereporturl         string
	Metadbcommandurl              string
	Metadbeventurl                string
	Metadbscheduleurl             string
	Metadbprovisionwatcherurl     string
	Metadbpingurl                 string
	ConsulProfilesActive          string
}

var configuration ConfigurationStruct = ConfigurationStruct{} //  Needs to be initialized before used

var (
	SupportSchedulerServiceName = "support-scheduler"
)
