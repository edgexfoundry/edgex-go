#################################
APIs - Core Services Core Command
#################################

======================
Architecture Reference
======================

For a description of the architecture, see :doc:`../Ch-Command` 

============
Introduction
============

EdgeX Foundry's Command microservice is a conduit for other services to trigger action on devices and sensors through their managing Device Services. The service provides an API to get the list of commands that can be issued for all devices or a single device. Commands are divided into two groups for each device:

* GET commands are issued to a device or sensor to get a current value for a particular attribute on the device, such as the current temperature provided by a thermostat sensor, or the on/off status of a light. 
* PUT commands are issued to a device or sensor to change the current state or status of a device or one of its attributes, such as setting the speed in RPMs of a motor, or setting the brightness of a dimmer light.

https://github.com/edgexfoundry/edgex-go/blob/master/core/command/raml/core-command.raml

.. _`Core Command API HTML Documentation`: file:core-command.html
..

`Core Command API HTML Documentation`_
