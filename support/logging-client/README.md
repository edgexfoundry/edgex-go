# README #
Support logging client library for the Go implementation of EdgeX microservices.  This project contains a logging client used to log messages.  Logging a message will send the log to the logging microservice, print the log to stdout and (optionally) write the log to a log file.

### What is this repository for? ###
* Logging client for logging messages

### Installation ###
This project uses glide for dependency management - https://glide.sh/
After installing glide, run the following commands to install the logging client:
```
go get github.com/edgexfoundry/support-logging-client-go
cd $GOPATH/src/github.com/edgexfoundry/support-logging-client-go
glide install
go install
```

### How To Use ###
To make logging calls, you need to have an instance of the LoggingClient struct.  Do not create manually.  Instead, call:
```
func NewClient(owningServiceName string, remoteUrl string) LoggingClient{...}
```
* owningServiceName - Name of your microservice (used in logs)
* remoteUrl - Full path to the logging service api
* 

You can optionally define a logging file for logs to be written to.  To do this, set the LogFilePath property on the LoggingClient instance.  The default path is an empty string which signifies that no log file will be created.

To log messages, call the respective commands (Info, Error, Debug, Warn)
