#######
Logging
#######

.. image:: EdgeX_SupportingServicesLogging.png

============
Introduction
============

Logging is critical for all modern software applications. Proper logging provides the users with the following benefits:

* Ability to monitor and understand what systems are doing
* Ability to understand how services interact with each other
* Problems are detected and fixed quickly
* Monitoring to foster performance improvements

The graphic shows the high-level design architecture of EdgeX Foundry including the Logging Service.

===========================
Minimum Product Feature Set
===========================

1. Provides a RESTful API for other microservices to request log entries with the following characteristics:

* The RESTful calls should be non-blocking, meaning calling services should fire logging requests without waiting for any response from the log service to achieve minimal impact to the speed and performance to the services.
* Support multiple logging levels, for example trace, debug, info, warn, error, fatal, and so forth.
* Each log entry should be associated with its originating service.

2. Provide RESTful APIs to query, clear, or prune log entries based on any combination of following parameters:

* Timestamp from
* Timestamp to
* Log level
* Originating service

3. Log entries should be persisted in either file or database, and the persistence storage should be managed at configurable levels. Querying via the REST API is only supported for deployments using database storage.
4. Take advantage of an existing logging framework internally and provide the “wrapper” for use by EdgeX Foundry
5. Follow applicable standards for logging where possible and not onerous to use on the gateway

==============================
High Level Design Architecture
==============================

.. image:: EdgeX_SupportingServicesLoggingArchitecture.png

The above diagram shows the high-level architecture for EdgeX Foundry Logging Service. Other microservices interact with EdgeX Foundry Logging Service through RESTful APIs to submit their logging requests, query historical logging, and remove historical logging. Internally, EdgeX Foundry's Logging Service utilizes the `GoKit logger <https://github.com/go-kit/kit/tree/master/log>`_ as its internal logging framework. Two configurable persistence options exist supported by EdgeX Foundry Logging Service: file or MongoDB.

========================
Configuration Properties
========================

+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
|   **Configuration**                                     |   **Default Value**                 |  **Dependencies**                                                         |
+=========================================================+=====================================+===========================================================================+
| **Entries in the Writable section of the configuration can be changed on the fly while the service is running if the service is running with the `--registry / -r` flag** |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Writable Persistence                                    | database                        \*  | "file" to save logging in file;                                           |
|                                                         |                                     | "database" to save logging in MongoDB                                     |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Writable LogLevel                                       | INFO                            \*  | Logs messages set to a level of "INFO" or higher                          |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| **The following keys represent the core service-level configuration settings**                                                                                            |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service MaxResultCount                                  | 50000                          \**  | Read data limit per invocation                                            |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service BootTimeout                                     | 300000                         \**  | Heart beat time in milliseconds                                           |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service StartupMsg                                      | Logging Service heart beat     \**  | Heart beat message                                                        |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service Port                                            | 48061                          \**  | Micro service port number                                                 |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service Host                                            | localhost                      \**  | Micro service host name                                                   |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service Protocol                                        | http                           \**  | Micro service host protocol                                               |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service ClientMonitor                                   | 15000                          \**  | The interval in milliseconds at which any service clients will            |
|                                                         |                                     | refresh their endpoint information from the service registry (Consul)     |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service CheckInterval                                   | 10s                            \**  | The interval in seconds at which the service registry (Consul) will       |
|                                                         |                                     | conduct a health check of this service.                                   |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Service Timeout                                         | 5000                           \**  | Specifies a timeout (in milliseconds) for handling requests               |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| **Following config only take effect when Writable.Persistence=file**                                                                                                      |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Logging File                                            | ./logs/edgex-support-logging.log    | File path to save logging entries                                         |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| **Following config only take effect when logging.persistence=database**                                                                                                   |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Databases Database Primary Username                     | [empty string]                 \**  | DB user name                                                              |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Databases Database Password                             | [empty string]                 \**  | DB password                                                               |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Databases Database Host                                 | localhost                      \**  | DB host name                                                              |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Databases Database Port                                 | 27017                          \**  | DB port number                                                            |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Databases Database Database                             | logging                        \**  | database or document store name                                           |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Databases Database Timeout                              | 5000                           \**  | DB connection timeout                                                     |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Databases Database Type                                 | mongodb                        \**  | DB type                                                                   |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| **Following config only take effect when connecting to the registry for configuraiton info**                                                                              |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Registry Host                                           | localhost                      \**  | Registry host name                                                        |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Registry Port                                           | 8500                           \**  | Registry port number                                                      |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Registry Type                                           | consul                         \**  | Registry implementation type                                              |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+


| \*means the configuration value can be changed on the fly if using a configuration registry (like Consul).
| \**means the configuration value can be changed but the service must be restarted.
| \***means the configuration value should NOT be changed.


====================================================
Logging Service Client Library for Go
====================================================

As the reference implementation of EdgeX Foundry microservices is written in Go, we provide a Client Library for Go so that Go-based microservices could directly switch their Loggers to use the EdgeX Foundry Logging Service.

The Go LoggingClient is part of the `go-mod-core-contracts module <https://github.com/edgexfoundry/go-mod-core-contracts>`_. This module can be imported into your project by including a reference in your go.mod. You can either do this manually or by executing "go get github.com/edgexfoundry/go-mod-core-contracts" from your project directory will add a reference to the latest tagged version of the module.

After that, simply import "github.com/edgexfoundry/go-mod-core-contracts/clients/logger" into a given package where your functionality will be implemented. Declare a variable or type member as logger.LoggingClient and it's ready for use.

::

    package main

    import "github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

    func main() {
        client := logger.LoggingClient

        //LoggingClient is now ready for use. A method is exposed for each LogLevel
        client.Trace("some info")
        client.Debug("some info")
        client.Info("some info")
        client.Warn("some info")
        client.Error("some info")
    }

::

Log statements will only be written to the log if they match or exceed the minimum LogLevel set in the configuration (described above). This setting can be changed on the fly without restarting the service to help with real-time troubleshooting.

Log statements are currently output in a simple key/value format. For example:

::

    level=INFO ts=2019-05-16T22:23:44.424176Z app=edgex-support-notifications source=cleanup.go:32 msg="Cleaning up of notifications and transmissions"

::

Everything up to the "msg" key is handled by the logging infrastructure. You get the log level, timestamp, service name and the location in the source code of the logging statement for free with every method invocation on the LoggingClient. The "msg" key's value is the first parameter passed to one of the Logging Client methods shown above. So to extend the usage example a bit, the above calls would result in something like:

::

    level=INFO ts=2019-05-16T22:23:44.424176Z app=logging-demo source=main.go:11 msg="some info"

::

You can add as many custom key/value pairs as you like by simply adding them to the method call:

::

    client.Info("some info","key1","abc","key2","def")

::

This would result in:

::

    level=INFO ts=2019-05-16T22:23:44.424176Z app=logging-demo source=main.go:11 msg="some info" key1=abc key2=def

::

Quotes are only put around values that contain spaces.

==================
EdgeX Logging Keys
==================
Within the Edgex Go reference implementation, log entries are currently written as a set of key/value pairs. We may change this later to be more of a struct type than can be formatted according to the user’s requirements (JSON, XML, system, etc). In that case, the targeted struct should contain properties that support the keys utilized by the system and described below.

+-----------------------------------------------+---------------------------------------------------------------------------------------+
|   **Key**                                     |   **Intent*                                                                           |
+===============================================+=======================================================================================+
| level                                         | Indicates the log level of the individual log entry (INFO, DEBUG, ERROR, etc)         |
+-----------------------------------------------+---------------------------------------------------------------------------------------+
| ts                                            | The timestamp of the log entry, recorded in UTC                                       |
+-----------------------------------------------+---------------------------------------------------------------------------------------+
| app                                           | This should contain the service key of the service writing the log entry              |
+-----------------------------------------------+---------------------------------------------------------------------------------------+
| source                                        | The file and line number where the log entry was written                              |
+-----------------------------------------------+---------------------------------------------------------------------------------------+
| msg                                           | A field for custom information accompanying the log entry. You do not need to         |
|                                               | specify this explicitly as it is the first parameter when calling one of the          |
|                                               | LoggingClient’s functions.                                                            |
+-----------------------------------------------+---------------------------------------------------------------------------------------+
| correlation-id                                | Records the correlation-id header value that is scoped to a given request.            |
|                                               | It has two sub-ordinate, associated fields (see below).                               |
+-----------------------------------------------+---------------------------------------------------------------------------------------+
| correlation-id path                           | This field records the API route being requested and is utilized when the             |
|                                               | service begins handling a request.                                                    |
|                                               | \* Example: path=/api/v1/event                                                        |
|                                               | When beginning the request handling, by convention set “msg” to “Begin request”.      |
+-----------------------------------------------+---------------------------------------------------------------------------------------+
| correlation-id duration                       | This field records the amount of time taken to handle a given request.                |
|                                               | When completing the request handling, by convention set “msg” to “Response complete”. |
+-----------------------------------------------+---------------------------------------------------------------------------------------+

Additional keys can be added as need warrants. This document should be kept updated to reflect their inclusion and purpose.





