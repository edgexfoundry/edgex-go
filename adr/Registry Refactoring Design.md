# Registry Refactoring Design

### Introduction

Currently the `Registry Client` in `go-mod-registry` module provides **Service Configuration** and **Service Registration** functionality. The goal of this design is to refactor the `go-mod-registry` module for separation of concerns. The **Service Registry** functionality will stay in the `go-mod-registry` module and the **Service Configuration** functionality will be separated out into a new `go-mod-configuration` module. This allows for implementations for deferent providers for each, another aspect of separation of concerns.

### Provider Connection information

An aspect of using the current `Registry Client` is "*Where do the services get the `Registry Provider` connection information?*" Currently all services either pull this connection information from the local configuration file or from the `edgex_registry` environment variable. Device Services also have the option to specify this connection information on the command line. With the refactoring for separation of concerns, this issue changes to "*Where do the services get the `Configuration Provider` connection information?*"

There have been concerns voiced by some in the EdgeX community that storing this `Configuration Provider` connection information in the configuration which ultimately is provided by that provider is not the right design.

This design proposes that all services will use the command line option approach with the ability to override with an environment variable. The  `Configuration Provider` information will **<u>not</u>** be stored in each service's local configuration file. The `edgex_registry` environment variable will be deprecated. The `Registry Provider` connection information will continue to be stored in each service's configuration either locally or from the`Configuration Provider` same as all other Edgex Client and Database connection information. An additional property will be added the the Registry Provider information to indicate if it is `Enabled`.

### Command line option changes

The new `-cp/-configProvider` command line option will be added to each service which will have a value specified using the format `{type}.{protocol}://{host}:{port}` e.g `consul.http://localhost:8500`.  This new command line option will be overridden by the `edgex_configuration_provider` environment variable when it is set. This environment variable's value has the same format as the command line option value. 

If **no value** is provided to the `-cp/-configProvider` option, i.e. just `-cp`, and no environment variable override is specified, the default value of `consul.http://localhost:8500` will be used. 

if `-cp/-configProvider` **not used** and no environment variable override is specified the local configuration file is used, as is it now.

All services will log the `Configuration Provider` connection information that is used.

The existing `-r/-registry` command line option will be removed from all services. 

### Bootstrap Changes

All services in the edgex-go mono repo use the new common bootstrap functionality.  The plan is to move this code to a go module for the Device Service and App Functions SDKs to also use. The current bootstrap modules `pkg/bootstrap/configuration/registry.go` and `pkg/bootstrap/container/registry.go` will be refactored to use the new `Configuration Client` and be renamed appropriately. New bootstrap modules will be created for using the revised version of `Registry Client` .  The current use of `useRegistry` and `registryClient` for service configuration will be change to appropriate names for using the new `Configuration Client`.  The current use of `useRegistry` and `registryClient` for service registration will be retained for service registration. Call to the new Unregister() API will be added to shutdown code for all services.

### Config-Seed Changes

The `conf-seed` service will have similar changes for specifying the `Configuration Provider` connection information since it doesn't use the common bootstrap package. Beyond that it will have minor changes for switching to using the `Configuration Client` interface, which will just be imports and appropriate name refactoring.

### Config Endpoint Changes

Since the `Configuration Provider` connection information will no longer be in the service's configuration struct, the `config` endpoint processing will be modified to add the `Configuration Provider` connection information to the resulting JSON create from service's configuration.

### Client Interfaces changes

##### Current Registry Client

This following is the current `Registry Client` Interface

```go
type Client interface {
	Register() error
	HasConfiguration() (bool, error)
	PutConfigurationToml(configuration *toml.Tree, overwrite bool) error
	PutConfiguration(configStruct interface{}, overwrite bool) error
	GetConfiguration(configStruct interface{}) (interface{}, error)
	WatchForChanges(updateChannel chan<- interface{}, errorChannel chan<- error, 
                    configuration interface{}, waitKey string)
	IsAlive() bool
	ConfigurationValueExists(name string) (bool, error)
	GetConfigurationValue(name string) ([]byte, error)
	PutConfigurationValue(name string, value []byte) error
	GetServiceEndpoint(serviceId string) (types.ServiceEndpoint, error)
	IsServiceAvailable(serviceId string) error
}
```

##### New Configuration Client

This following is the new `Configuration Client` Interface which contains the  **Service Configuration** specific portion from the above current `Registry Client`.

```go
type Client interface {
	HasConfiguration() (bool, error)
	PutConfigurationToml(configuration *toml.Tree, overwrite bool) error
	PutConfiguration(configStruct interface{}, overwrite bool) error
	GetConfiguration(configStruct interface{}) (interface{}, error)
	WatchForChanges(updateChannel chan<- interface{}, errorChannel chan<- error,
                    configuration interface{}, waitKey string)
	IsAlive() bool
	ConfigurationValueExists(name string) (bool, error)
	GetConfigurationValue(name string) ([]byte, error)
	PutConfigurationValue(name string, value []byte) error
}
```

##### Revised Registry Client

This following is the revised `Registry Client` Interface, which contains the **Service Registry** specific portion from the above current `Registry Client`. The `UnRegister()` API has been added per issue [#20](https://github.com/edgexfoundry/go-mod-registry/issues/20)

```go
type Client interface {
	Register() error
	UnRegister() error
	IsAlive() bool
	GetServiceEndpoint(serviceId string) (types.ServiceEndpoint, error)
	IsServiceAvailable(serviceId string) error
}
```

### Client Configuration Structs

##### Current Registry Client Config

The following is the current `struct` used to configure the current `Registry Client`

```go
type Config struct {
	Protocol string
	Host string
	Port int
	Type string
	Stem string
	ServiceKey string
	ServiceHost string
	ServicePort int
	ServiceProtocol string
	CheckRoute string
	CheckInterval string
}
```

##### New Configuration Client Config

The following is the new `struct` the will be used to configure the new `Configuration Client` from the command line option or environment variable values. The Service Registry portion has been removed from the above existing `Registry Client Config`

```go
type Config struct {
	Protocol string
	Host string
	Port int
	Type string
	BasePath string
	ServiceKey string
}
```

##### New Registry Client Config

The following is the revised `struct` the will be used to configure the new `Registry Client` from the  information in the service's configuration. This is mostly unchanged from the existing `Registry Client Config`, except that the `Stem` for configuration has been removed

```go
type Config struct {
	Protocol string
	Host string
	Port int
	Type string
	ServiceKey string
	ServiceHost string
	ServicePort int
	ServiceProtocol string
	CheckRoute string
	CheckInterval string
}
```

### Provider Implementations

The current `Consul` implementation of the `Registry Client` will be split up into implementations for the new `Configuration Client` in the new `go-mod-configuration` module and the revised `Registry Client` in the existing `go-mod-registry` module.