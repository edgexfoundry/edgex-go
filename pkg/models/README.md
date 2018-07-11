# README #
This is the public models package for the Go implementation of the EdgeX microservices. These models constitute data representation when calling the external APIs of the EdgeX platform.

### What is this repository for? ###
* Public models for EdgeX microservices

### How to use? ###
Import the project using `go get https://github.com/edgexfoundry/core-domain-go`

You can also use [glide](https://glide.sh) to keep this library up to date by adding the following to your glide.yaml
```
- package: github.com/edgexfoundry/edgex-go/pkg/models
  subpackages:
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
