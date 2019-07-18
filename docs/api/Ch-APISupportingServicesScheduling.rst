#######################################
APIs - Supporting Services - Scheduling
#######################################

======================
Architecture Reference
======================

For a description of the architecture, see :doc:`../Ch-Scheduling` 

============
Introduction
============

The following is and example of the EdgeX Foundry Scheduler Service RESTful API.

https://github.com/edgexfoundry/edgex-go/blob/master/api/raml/support-scheduling.raml

.. _`Scheduling API HTML Documentation`: support-scheduler.html
..

`Scheduling API HTML Documentation`_


===========
Description
===========

**Scheduler Service** - a service that can be used to schedule invocation of a URL. Requires the use of interval(s), and interval action(s).

**Interval(s)**
    * **name** - unique name of the service.
    * **start** - identifies when the operation starts. Expressed in ISO 8601 YYYYMMDD'T'HHmmss format. Empty means now.
    * **end** - identifies when the operation ends. Expressed in ISO 8601 YYYYMMDD'T'HHmmss format. Empty means never
    * **frequency** - identifies the interval between invocations. Expressed in ISO 8601 PxYxMxDTxHxMxS format. Empty means no frequency.

**Interval Action(s)**
    * **name** - unique name of the interval action.
    * **interval** - unique name of an existing interval.
    * **target** - the recipient of the interval action (ergo service or name).
    * **protocol** - the protocol type to be used to contact the target. (example HTTP).
    * **httpMethod** - HTTP protocol verb.
    * **address** - the endpoint server host.
    * **port** - the desired port.
    * **path** - the api path which will be acted on.
    * **parameters** - (optional) parameters which will be included in the BODY tag for HttpMethods. Any parameters that should be provided to the call, e.g. {"milliseconds":86400000}


========
Examples
========

Create an interval upon which the scheduler will operate
::

curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
   "name": "midnight",
   "start": "20180101T000000",
   "frequency": "P1D"}' "http://localhost:48081/api/v1/interval"


Example of a second interval which will run every 20 seconds
::

curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
   "name": "every20s",
   "start":"20000101T000000",
   "end":"",
   "frequency":"PT20S"}' "http://localhost:48081/api/v1/interval"

Create an interval action that will invoke the interval action (drive the scrubber) in core-data
::

curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
    "name": "scrub-pushed-events",
    "interval": "midnight",
    "target": "core-data",
    "protocol": "http",
    "httpMethod": "DELETE",
    "address": "localhost",
    "port": 48080,
    "path": "/api/v1/event/scrub"}' "http://localhost:48085/api/v1/intervalaction"


This is a Random-Boolean-Device which created by edgex-device-virtual that connects every 20 seconds.
::

curl -X POST -H "Content-Type: application/json" -d '{
    "name": "put-action",
    "interval": "every20s",
    "target": "edgex-device-modbus",
    "protocol": "http",
    "httpMethod": "PUT",
    "address": "localhost",
    "port": 49990,
    "path":"/api/v1/device/name/Random-Boolean-Device/RandomValue_Bool",
    "parameters": "{\"RandomValue_Bool\": \"true\",\"EnableRandomization_Bool\": \"true\"}"
}'  "http://localhost:48085/api/v1/intervalaction"