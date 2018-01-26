# README #
Consul client library for the Go implementation of EdgeX microservices.  This project contains functions for initializing a connection to the Consul service, registering health checks and pulling key/value pairs from the consul service.

### What is this repository for? ###
* Initialize connection to Consul
* Pull key/value pairs into a configuration struct

### Installation ###
consul-client-go uses the glide vendoring tool - https://glide.sh/
To pull the dependecies into your workspace and build the project, make sure you are in the project directory and run:
```
go get github.com/edgexfoundry/consul-client-go
cd $GOPATH/src/github.com/edgexfoundry/consul-client-go
glide install
go install
```

### How to Use ###
This library is used by Go programs for interacting with the Consul service and requires that a Consul agent is running somewhere that the Consul client can connect to.  The host address and port of the Consul agent will be used to initialize a connection to the Consul service.

This library has 2 function calls for interacting with a Consul service:
```
func CheckKeyValuePairs(configurationStruct interface{}, applicationName string, profiles []string) error{ ... }
```
and
```
func ConsulInit(config ConsulConfig) error{ ... }
```
CheckKeyValuePairs
* configurationStruct is struct of kay/values
  * Variable names will become the key names in consul
  * Pass a pointer to the struct so the actual struct is updated
  * profiles is an optional list of strings for organization in Consul

ConsulInit
* Create a ConsulConfig object to initialize consul connection

Note: Look at the core microservices for examples on how to use the consul-client-go library
