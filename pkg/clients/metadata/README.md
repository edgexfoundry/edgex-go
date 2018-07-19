# README #
This package contains the metadata client written in the Go programming language.  The metadata client is used by Go services or other Go code to communicate with the EdgeX core-metadata microservice (regardless of underlying implemenation type) by sending REST requests to the service's API endpoints.

### How To Use ###
To use the core-metadata client package you first need to import the library into your project:
```
import "github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
```
As an example of use, to find a device using the Metadata client, first create a new device client (see core-data init.go)
```
mdc = metadata.NewDeviceClient(params, types.Endpoint{})
```
And then use the device client to located a device by Device struct (see core-data event.go)
```
_, err := mdc.CheckForDevice(device)
```
