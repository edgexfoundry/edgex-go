# README #
This package contains the logging client written in the Go programming language.  The logging client is used by Go services or other Go code to communicate with the EdgeX support-logging microservice (regardless of underlying implemenation type) by sending REST requests to the service's API endpoints.

### How To Use ###
To use the support-logging client package you first need to import the library into your project:
```
import "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
```
To send a log message to the centralized logging service, you first need to get a LoggingClient and then you can send logging messages into the service (indicating the level with the various log function call.
```
  logTarget := setLoggingTarget(*configuration)
	loggingClient = logger.NewClient(internal.CoreDataServiceKey, configuration.EnableRemoteLogging, logTarget)

  loggingClient.Info(consulMsg)
	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.CoreDataServiceKey, edgex.Version))
```
Log messages can be logged as Info, Error, Debug, or Warn.
