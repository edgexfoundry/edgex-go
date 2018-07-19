# README #
This package contains the scheduler client written in the Go programming language.  The scheduler client is used by Go services or other Go code to communicate with the EdgeX support-scheduler microservice (regardless of underlying implemenation type) by sending REST requests to the service's API endpoints.

### How To Use ###
To use the support-scheduler client package you first need to import the library into your project:
```
import "github.com/edgexfoundry/edgex-go/pkg/clients/scheduler"
```
To work with schedules you first need to get a ScheduleClient:
```
scheduleClient := scheduler.SchedulerClient{
    SchedulerServiceHost : "localhost",
    SchedulerServicePort : 48081,
    OwningService : "My Service Name"
}
```

This will create a client to hit the scheduler endpoint.  You can then post a schedule by creating a Schedule struct and call:

```
schedule := models.Schedule{
	Name : "midnight",
	Start : ""20000101T000000"",
	End : "",
	Frequency : "P1D",
	Cron : "This is a description",
	RunOnce : ""
}
err := scheduleClient.AddSchedule(schedule)
```

And you can post a schedule event by creating a ScheduleEvent object and call:

```
scheduleEvent := models.ScheduleEvent{
	Name:       "pushed events",
	Parameters: "",
	Service:    "notifications",
	Schedule:   "testSchedule",
	Addressable: models.Addressable{
		Name:     "MQTT",
		Protocol: "MQTT",
	},
}

err := scheduleClient.AddScheduleEvent(scheduleEvent)
```
