# README #
This is Core Domain library for the Go implementation of the EdgeX microservices. This project contains the models that the other microservices use to pass around object data through http requests and to put object data into a database.

### What is this repository for? ###
* Domain objects for EdgeX microservices

### How to use? ###
Import the project using `go get https://github.com/edgexfoundry/core-domain-go`

You can also use [glide](https://glide.sh) to keep this library up to date by adding the following to your glide.yaml
```
- package: github.com/edgexfoundry/core-domain-go
  subpackages:
  - models
  - enums
```
### Models that this Project Contains ###
This Project contains:
- Action
- ActionType
- Addressable
- AdminState
- BaseObject
- CallBackAlert
- Command
- CommandResponse
- DescribedObject
- Device
- DeviceObject
- DeviceProfile
- DeviceReport
- DeviceService
- Event
- Get
- NotifyAction
- OperatingState
- ProfileProperty
- ProfileResource
- ProfileSource
- PropertyValue
- Protocol
- ProvisionWatcher
- Put
- Reading
- ResourceOperation
- Response
- Schedule
- ScheduleEvent
- Service
- Units
- ValueDescriptor
