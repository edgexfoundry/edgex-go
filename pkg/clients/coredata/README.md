# README #
This package contains the core data client written in the Go programming language.  The core data client is used by Go services or other Go code to communicate with the EdgeX core-data microservice (regardless of underlying implemenation type) by sending REST requests to the service's API endpoints.

### How To Use ###
To use the core-data client package you first need to import the library into your project:
```
import "github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
```
As an example of use, to find a Value Descriptor using the Core Data client, first create a new device client 
```
vdc := NewValueDescriptorClient(params, types.Endpoint{})
```
And then use the client to get all value descriptors
```
vdc.ValueDescriptors()
```
