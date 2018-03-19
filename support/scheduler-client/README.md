# README #
This repository is for the scheduler client for EdgeXFoundry written in the Go programming language.  The scheduler client is used to communicate with the scheduler micro service by sending REST requests to the service's API endpoints.

### What is this repository for? ###
* Client library for interacting with the scheduler microservice

### Installation ###
This project does not have any external dependencies.  To install, simply run:

```
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go/support/scheduler-client
go install
```

To test, simple run:

```
go test
```

### How To Use ###
To use the support-scheduler-client library you first need to import the library into your project:

```
import "github.com/edgexfoundry/edgex-go/support/scheduler-client"
```

To send schedule, schedule event you first need to create a SchedulerClient object:

```
scheduleClient := scheduler.SchedulerClient{
    RemoteScheduleUrl : "http://localhost:48081/api/v1/schedule",
    RemoteScheduleEventUrl : "http://localhost:48081/api/v1/scheduleevent",
    OwningService : "My Service Name"
}
```

This will create a client to hit the scheduler endpoint running on localhost.  You can then post a schedule by creating a Schedule object and call:

```
schedule := models.Schedule{
	Name : "midnight",
	Start : ""20000101T000000"",
	End : "",
	Frequency : "P1D",
	Cron : "This is a description",
	RunOnce : ""
}
err := scheduleClient.SendSchedule(schedule)
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

err := scheduleClient.SendScheduleEvent(scheduleEvent)
```


more API details please see the [EdgeXFoundry's API Wiki - APIs--Supporting Services--Scheduling](https://wiki.edgexfoundry.org/display/FA/APIs--Supporting+Services--Scheduling)