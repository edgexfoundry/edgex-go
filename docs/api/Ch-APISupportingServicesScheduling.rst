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

The following API RESTful Web Services for EdgeX Foundry at this time shows the older development name of "fuse."  The EdgeX Foundry name in the software will be updated soon to "edgexfoundry." 

https://github.com/edgexfoundry/edgex-go/blob/master/api/raml/support-scheduling.raml

.. _`Scheduling API HTML Documentation`: support-scheduler.html
..

`Scheduling API HTML Documentation`_


===========
Description
===========

**Scheduler Service** - a service that can be used to schedule invocation of a URL. Requires the use of addressables, schedules, and schedule events.

**Addressables**

    identify the called service

**Schedules**

    name - unique name of the service.
    start - identifies when the operation starts. Expressed in ISO 8601 YYYYMMDD'T'HHmmss format. Empty means now.
    end - identifies when the operation ends. Expressed in ISO 8601 YYYYMMDD'T'HHmmss format. Empty means never
    frequency - identifies the interval between invocations. Expressed in ISO 8601 PxYxMxDTxHxMxS format. Empty means no frequency.

**Schedule Events** name - unique name of the event. addressable - address information of the service to invoke parameters - any parameters that should be provided to the call, e.g. {"milliseconds":86400000} service - identifies the service that should execute the event, e.g. fuse-support-scheduler * schedule - associates the event to a schedule

========
Examples
========

Create an Addressable for the service requiring invocation

::

   curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{ 
   "origin":1234567890,
   "name":"pushed events",
   "protocol":"HTTP",
   "address":"localhost",
   "port":48080,
   "path":"/api/v1/event/scrub",
   "publisher":null,
   "user":null,
   "password":null,
   "topic":null }' "http://localhost:48081/api/v1/addressable"

::

   curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{ 
   "origin":1234567890,
   "name":"aged events",
   "protocol":"HTTP",
   "address":"localhost",
   "port":48080,
   "path":"/api/v1/event/removeold/age/604800000",
   "publisher":null,
   "user":null,
   "password":null,
   "topic":null }' "http://localhost:48081/api/v1/addressable"

Create a schedule upon which the scheduler will operate

::

   curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{ 
   "origin":1234567890,
   "name":"midnight",
   "start":"20000101T000000",
   "end":"",
   "frequency":"P1D"}' "http://localhost:48081/api/v1/schedule"

Create an event that will use the schedule and invoke the addressable (drive the scrubber)

::

   curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{ 
   "origin":1234567890,
   "name":"pushed events",
   "addressable": { "name" : "pushed events" },
   "parameters": null,
   "service" : "fuse-support-scheduler",
   "schedule":"midnight"}' "http://localhost:48081/api/v1/scheduleevent"

::

   curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{ 
   "origin":1234567890,
   "name":"aged events",
   "addressable": { "name" : "aged events" },
   "parameters": null,
   "service" : "fuse-support-scheduler",
   "schedule":"midnight"}' "http://localhost:48081/api/v1/scheduleevent"

For testing update the midnight service to be every 60 seconds

::

   curl -X PUT -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{ 
   "name":"midnight",
   "frequency":"PT60S"}' "http://localhost:48081/api/v1/schedule"

