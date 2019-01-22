# README #
Registry client library for the Go implementation of EdgeX micro services.  This project contains the abstract Registry interface and an implementation for Consul.
These interface functions initialize a connection to the Registry service, registering the service for discovery and  health checks, push and pull configuration values to/from the Registry service and pull dependent service endpoint information and status.

### What is this repository for? ###
* Initialize connection to a Registry service
* Register the service with the Registry service for discovery, health status and health check callback
* Push a service's configuration in to the Registry
* Pull service's configuration from the Registry into its configuration struct
* Pull service endpoint information from the Registry for dependent services.
* Check the health status of dependent services via the Registry.

### Installation ###
TBD once Go Modules is in full use.

### How to Use ###
This library is used by Go programs for interacting with the configured Registry service (i.e. Consul) and requires that a Registry service be running somewhere that the Registry Client can connect.  The Registry service connection information as well as which registry implementation to use is stored in the service's toml configuration as:

        [Registry]
        Host = 'localhost'
        Port = 8500
        Type = 'consul'

The following code snippets demonstrate how a service uses this Registry module to register, load configuration, listen to for configuration updates and to get dependent service endpoint information.

This code snippet shows how to connect to the Registry and register the service for discovery and health checks. Note that the expected health check callback URL path is "/api/v1/ping" which your service must implement. 
```
func connectToRegistry(conf *ConfigurationStruct) (error) {
	var err error

	registryClient, err := registry.NewRegistryClient(conf.Registry, &conf.Service, internal.CoreDataServiceKey)
	if err != nil {
		return fmt.Errorf("connection to Registry could not be made: %v", err.Error())
	}

	// Register the service with Registry
	err = registryClient.Register()
	if err != nil {
		return fmt.Errorf("could not register service with Registry: %v", err.Error())
	}

	return nil
}
```
This code snippet shows how to load the service's configuration from the Registry after connecting and registering above.
```
    if useRegistry {
      err = connectToRegistry(configuration)
      if err != nil {
        return nil, err
      }

      rawConfig, err := registry.Client.GetConfiguration(configuration)
      if err != nil {
        return configuration, fmt.Errorf("could not get configuration from Registry: %v", err.Error())
      }

      actual, ok := rawConfig.(*ConfigurationStruct)
      if !ok {
        return configuration, fmt.Errorf("configuration from Registry failed type check")
      }

      configuration = actual
    }
```
This code snippet shows how to listen for configuration changes from the Registry after connecting and registering above.
Note the reference to the RegistryClient was saved in registry.Client and used here.
```
    registry.Client.WatchForChanges(updateChannel, errChannel, &WritableInfo{}, internal.WritableKey)

    for {
      select {
      case ex := <-errChannel:
        LoggingClient.Error(ex.Error())

      case raw, ok := <-updateChannel:
        if !ok {
          return
        }

        actual, ok := raw.(*WritableInfo)
        if !ok {
          LoggingClient.Error("listenForConfigChanges() type check failed")
          return
        }

        Configuration.Writable = *actual

        LoggingClient.Info("Writeable configuration has been updated. Setting log level to " +
          Configuration.Writable.LogLevel)
        LoggingClient.SetLogLevel(Configuration.Writable.LogLevel)
      }
    }
```
This code snippet shows how to get dependent service endpoint information and check status of the dependent service.
```
    ...
    if registry.Client != nil {
        endpoint, err = registry.Client.GetServiceEndpoint(params.ServiceKey)
        url := fmt.Sprintf("http://%s:%v%s", endpoint.Address, endpoint.Port, params.Path)
        ...
        if registry.Client.IsServiceAvailable(params.ServiceKey) {
           ...
        }
    } 
    ...
```