# README #
This package contains the logging client written in the Go programming language.  The logging client is used by Go services or other Go code to log messages to STDOUT.

### How To Use ###
To use the logging client package you first need to import the library into your project:
```
import "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logging"
```
To send a log message to STDOUT, you first need to create a LoggingClient with desired Log Level and then you can send log messages (indicating the log level of the message using one of the various log function calls).
```
loggingClient = logger.NewClient(internal.CoreDataServiceKey,configuration.Writable.LogLevel) 

loggingClient.Info("Something interesting")
loggingClient.Infof("Starting %s %s ", internal.CoreDataServiceKey, edgex.Version)
loggingClient.Errorf("Something bad happened: %s", err.Error())
```
Log messages can be logged as Info, Debug, Trace, Warn, or Error
