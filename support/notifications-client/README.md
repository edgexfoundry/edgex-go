# README #
This repository is for the notifications client for EdgeXFoundry written in the Go programming language.  The notifications client is used to communicate with the notifications micro service by sending REST requests to the service's API endpoints.

### What is this repository for? ###
* Client library for interacting with the notifications microservice

### Installation ###
This project does not have any external dependencies.  To install, simply run:
```
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go/support/notifications-client
go install
```

### How To Use ###
To use the support-notifications-client library you first need to import the library into your project:
```
import "github.com/edgexfoundry/edgex-go/support/notifications-client"
```
To send a notification you first need to create a NotificationsClient object:
```
nc := notifications.NotificationsClient{
    NotificationsServiceHost : "localhost",
    NotificationsServicePort : 48060,
    OwningService : "My Service Name",
}
```
This will create a client to hit the notifications endpoint running on localhost.  You can then post a notifications by creating a Notification object and call:
```
n := notifications.Notification{
	Sender : "Microservice Name",
	Category : notifications.SW_HEALTH,
	Severity : notifications.NORMAL,
	Content : "This is a notification",
	Description : "This is a description",
	Status : notifications.NEW,
	Labels : []string{"Label One", Label Two"},
}
err := nc.SendNotification(n)
```
This will send the notification to the notifications service.
