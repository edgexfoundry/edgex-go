#############################
System Management Agent (SMA)
#############################

============
Introduction
============

While the SMA serves several purposes, it is to be considered, first and foremost, as the single connection point of management control for an EdgeX instance. As such, the API calls related to system management are defined so as to interface with the SMA.

=====================
Examples of API Calls
=====================

To get an appreciation for some SMA API calls in action, it will be instructive to look at what responses the SMA provides to the caller, for the respective calls (Notice, too, the error messages returned by the SMA, should it encounter a problem). Thus, consider the following calls that the SMA handles:

* `Metrics of a service`_
* `Configuration of a service`_
* `Start a service`_
* `Stop a service`_
* `Restart a service`_
* `Health check on a service`_

Let's look at the preceding calls (aka requests), one-by-one, in the following sections.

--------------------
Metrics of a service
--------------------

Example request:
/api/v1/metrics/edgex-core-command,edgex-core-data

Corresponding response, in JSON format:

.. code-block:: json

    {
       "Metrics":{
          "edgex-core-command":{
             "CpuBusyAvg":2.224995150836366,
             "Memory":{
                "Alloc":1403648,
                "Frees":1504,
                "LiveObjects":18280,
                "Mallocs":19784,
                "Sys":71891192,
                "TotalAlloc":1403648
             }
          },
          "edgex-core-data":{
             "CpuBusyAvg":2.854720153816541,
             "Memory":{
                "Alloc":929080,
                "Frees":1453,
                "LiveObjects":7700,
                "Mallocs":9153,
                "Sys":70451200,
                "TotalAlloc":929080
             }
          }
       }
    }

--------------------------
Configuration of a service
--------------------------

Example request:
/api/v1/config/device-simple,edgex-core-data

Corresponding response, in JSON format:

.. code-block:: json

    {
        "Configuration": {
            "device-simple": "device-simple service is not registered. Might not have started... ",
            "edgex-core-data": {
                "Clients": {
                    "Logging": {
                        "Host": "localhost",
                        "Port": 48061,
                        "Protocol": "http"
                    },
                    "Metadata": {
                        "Host": "localhost",
                        "Port": 48081,
                        "Protocol": "http"
                    }
                },
                "Databases": {
                    "Primary": {
                        "Host": "localhost",
                        "Name": "coredata",
                        "Password": "",
                        "Port": 27017,
                        "Timeout": 5000,
                        "Type": "mongodb",
                        "Username": ""
                    }
                },
                "Logging": {
                    "EnableRemote": false,
                    "File": "./logs/edgex-core-data.log"
                },
                "MessageQueue": {
                    "Host": "*",
                    "Port": 5563,
                    "Protocol": "tcp",
                    "Type": "zero"
                },
                "Registry": {
                    "Host": "localhost",
                    "Port": 8500,
                    "Type": "consul"
                },
                "Service": {
                    "BootTimeout": 30000,
                    "CheckInterval": "10s",
                    "ClientMonitor": 15000,
                    "Host": "localhost",
                    "Port": 48080,
                    "Protocol": "http",
                    "MaxResultCount": 50000,
                    "StartupMsg": "This is the Core Data Microservice",
                    "Timeout": 5000
                },
                "Writable": {
                    "DeviceUpdateLastConnected": false,
                    "LogLevel": "INFO",
                    "MetaDataCheck": false,
                    "PersistData": true,
                    "ServiceUpdateLastConnected": false,
                    "ValidateCheck": false
                }
            }
        }
    }

---------------
Start a service
---------------

Example request:
/api/v1/operation

Example (POST) body accompanying the "start" request:

.. code-block:: json

    {
       "action":"start",
       "services":[
          "edgex-core-data",
          "edgex-export-distro"
       ],
       "params":[
       	"graceful"
       	]
    }

Corresponding response, in JSON format, on success:
"Done. Started the requested services."

Corresponding response, in JSON format, on failure:
"HTTP 500 - Internal Server Error"

--------------
Stop a service
--------------

Example request:
/api/v1/operation

Example (POST) body accompanying the "stop" request:

.. code-block:: json

    {
       "action":"stop",
       "services":[
          "edgex-support-notifications"
       ],
       "params":[
       	"graceful"
       	]
    }

Corresponding response, in JSON format, on success:
"Done. Stopped the requested service."

Corresponding response, in JSON format, on failure:
"HTTP 500 - Internal Server Error"

-----------------
Restart a service
-----------------

Example request:
/api/v1/operation

Example (POST) body accompanying the "restart" request:

.. code-block:: json

    {
       "action":"restart",
       "services":[
          "edgex-support-notifications",
          "edgex-core-data",
          "edgex-export-distro"

       ],
       "params":[
       	"graceful"
       	]
    }

Corresponding response, in JSON format, on success:
"Done. Restarted the requested services."

Corresponding response, in JSON format, on failure:
"HTTP 500 - Internal Server Error"

-------------------------
Health check on a service
-------------------------

Example request:
/api/v1/health/device-simple,edgex-core-data,support-notifications

Corresponding response, in JSON format:

.. code-block:: json

    {
        "device-simple": "device-simple service is not registered. Might not have started... ",
        "edgex-core-data": true,
        "support-notifications": true
    }
