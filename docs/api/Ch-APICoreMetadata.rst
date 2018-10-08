###################################
APIs - Core Services -Core Metadata
###################################

======================
Architecture Reference
======================

For a description of the architecture, see :doc:`../Ch-Metadata` 

============
Introduction
============

The Metadata microservice includes the device/sensor metadata database and APIs to expose the database to other services. In particular, the device provisioning service deposits and manages device metadata through this service. This service may also hold and manage other configuration metadata used by other services on the gateway such as clean up schedules, hardware configuration (Wi-Fi connection info, MQTT queues, and so forth). Non-device metadata may need to be held in a different database and/or managed by another service–depending upon implementation.

https://github.com/edgexfoundry/edgex-go/blob/master/api/raml/core-metadata.raml

.. _`Core Metadata API HTML Documentation`: core-metadata.html
..

`Core Metadata API HTML Documentation`_
