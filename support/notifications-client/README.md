# README #
This repository is for the notifications client for EdgeXFoundry written in the Go programming language.  The notifications client is used to communicate with the notifications micro service by sending REST requests to the service's API endpoints.

### What is this repository for? ###
* Client library for interacting with the notifications microservice

### Installation ###
This project does not have any external dependencies.  To install, simply run:
```
go get github.com/edgexfoundry/support-notifications-client-go
cd $GOPATH/src/github.com/edgexfoundry/support-notifications-client-go
go install
```

### How To Use ###
To use the support-notifications-client library you first need to import the library into your project:
```
import "github.com/edgexfoundry/support-notifications-client-go"
```
To send a notification you first need to create a NotificationsClient object:
```
nc := notifications.NotificationsClient{
    RemoteUrl : "http://localhost:48060/api/v1/notification",
    OwningService : "My Service Name"
}
```
This will create a client to hit the notifications endpoint running on localhost.  You can then post a notifications by creating a Notification object and call:
```
n := notifications.Notification{
	Sender : "Microservice Name",
	Category : notifications.SW_HEALTH,
	Severity : notifications.NORMAL,
	Content : "This is a notification",
	Description string : "This is a description",
	Status StatusEnum : notifications.NEW,
	Labels []string	: []string{"Label One", Label Two"}
}
err := nc.ReceiveNotification(n)
```
This will send the notification to the notifications client and return an error.
