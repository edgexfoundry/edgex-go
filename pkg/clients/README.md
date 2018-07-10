# README #
Core client library for the Go implementation of EdgeX microservices.  This project contains client libraries for interacting with Go core microservices.

### What is this repository for? ###
* Client libraries for interacting with the core microservices

### Installation ###
This project uses glide for dependency management - https://glide.sh/
After installing glide, run the following commands to install the core client libraries:
```
go get github.com/edgexfoundry/core-clients-go
cd $GOPATH/src/github.com/edgexfoundry/core-clients-go
glide install
go install ./coredataclients
go install ./metadataclients
```

### How To Use ###
To use the core client libraries you first need to import the libraries into your project:
```
import "github.com/edgexfoundry/core-clients-go/coredataclients"
import "github.com/edgexfoundry/core-clients-go/metadataclients"
```
Each API endpoint for the respective microservice has a separate client object that you need to create.  There are constructer methods for doing this which are passed the URL for the api endpoint.  For example, to create a client object for using the device API of metadata, do the following:
```
d := metadataclients.NewDeviceClient("http://localhost:48081/api/v1/device")
```
This will create a client to hit the device endpoint of metadata running on localhost.  You can then call methods like:
```
devices, err := d.Devices()
```
This will return a list of devices that are currently present on metadata
