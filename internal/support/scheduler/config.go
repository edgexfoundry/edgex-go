package scheduler

import "github.com/edgexfoundry/edgex-go/internal/pkg/config"

// Configuration V2 for the Support Scheduler Service
type ConfigurationStruct struct {
	ScheduleInterval int

	Clients   map[string]config.ClientInfo
	Logging   config.LoggingInfo
	Registry  config.RegistryInfo
	Service   config.ServiceInfo
	Schedules map[string]config.ScheduleInfo
	ScheduleEvents map[string]config.ScheduleEventInfo

}