################################
Using the Virtual Device Service
################################

Overview
========

The `Virtual Device Service GO <https://github.com/edgexfoundry/device-virtual-go>`_ can simulate different kinds of devices to generate Events and Readings to the Core Data Micro Service. Furthermore, users can send commands and get responses through the Command and Control Micro Service. The Virtual Device Service allows you to execute functional or performance tests without any real devices. This version of the Virtual Device Service is implemented based on `Device SDK GO <https://github.com/edgexfoundry/device-sdk-go>`_, and uses `ql <https://godoc.org/modernc.org/ql>`_ (an embedded SQL database engine) to simulate virtual resources.

.. image:: Virtual_DS.png
   :scale: 60%
   :alt: Virtual Device Service

Sequence Diagram
----------------

.. image:: VirtualSequence.png
   :scale: 60%
   :alt: Sequence Diagram

Virtual Resource Table Schema
=============================

.. csv-table::
  :header: "Column", "Type"
  :widths: 20, 10

  "DEVICE_NAME", "STRING"
  "COMMAND_NAME", "STRING"
  "DEVICE_RESOURCE_NAME", "STRING"
  "ENABLE_RANDOMIZATION", "BOOL"
  "DATA_TYPE", "STRING"
  "VALUE", "STRING"

How to Use
==========

The Virtual Device Service depends on the EdgeX Core Services. If you're going to download the source code and run the Virtual Device Service in dev mode, make sure that the EdgeX Core Services are up before starting the Virtual Device Service.

The Virtual Device Service currently contains four pre-defined devices (see the `configuration.toml <https://github.com/edgexfoundry/device-virtual-go/blob/master/cmd/res/configuration.toml>`_) as random value generators:

.. csv-table::
  :header: "Device Name", "Device Profile"
  :widths: 20, 20

  "Random-Boolean-Generator01", "`device.virtual.bool.yaml <https://github.com/edgexfoundry/device-virtual-go/blob/master/cmd/res/device.virtual.bool.yaml>`_"
  "Random-Float-Generator01", "`device.virtual.float.yaml <https://github.com/edgexfoundry/device-virtual-go/blob/master/cmd/res/device.virtual.float.yaml>`_"
  "Random-Integer-Generator01", "`device.virtual.int.yaml <https://github.com/edgexfoundry/device-virtual-go/blob/master/cmd/res/device.virtual.int.yaml>`_"
  "Random-UnsignedInteger-Generator01", "`device.virtual.uint.yaml <https://github.com/edgexfoundry/device-virtual-go/blob/master/cmd/res/device.virtual.uint.yaml>`_"

**Restricted:** Resource names are currently hard coded, if you need to use your own device profile, you must update the deviceResource field in the default device profiles. For example if your device profile needs a boolean resource, the deviceResource field must define the "EnableRandomization_Bool" and "RandomValue_Bool" resources. This restriction will be removed in the next dot release.

Acquire the executable commands information by inquiring the Core Command API:

* http://[host]:48082/api/v1/device/name/Random-Boolean-Generator01
* http://[host]:48082/api/v1/device/name/Random-Integer-Generator01
* http://[host]:48082/api/v1/device/name/Random-UnsignedInteger-Generator01
* http://[host]:48082/api/v1/device/name/Random-Float-Generator01

GET command example
-------------------

.. image:: VirtualGET.png
   :scale: 60%
   :alt: GET command

PUT command example - Assign a value to a resource
--------------------------------------------------

The value must be a valid value for the data type. For example, the minimum value of RandomValue_Int8 cannot be less than -128 and the maximum value cannot be greater than 127.

.. image:: VirtualPUT_1.png
   :scale: 60%
   :alt: PUT command: Assign a value to a resource

PUT command example - Enable/Disable the randomization of the resource
----------------------------------------------------------------------

.. image:: VirtualPUT_2.png
   :scale: 60%
   :alt: PUT command: Enable/Disable the randomization of the resource

.. NOTE::

* The value of the resource's EnableRandomization property is simultaneously updated to false when sending a put command to assign a specified value to the resource
* The minimum and maximum values of the resource can be defined in the property value field of the Device Resource model, for example::

      deviceResources:
       -
         name: "RandomValue_Int8"
         description: "Generate random int8 value"
         properties:
           value:
             { type: "Int8", readWrite: "R", minimum: "-100", maximum: "100", defaultValue: "0" }
           units:
             { type: "String", readWrite: "R", defaultValue: "random int8 value" }

Manipulate Virtual Resources Using the command ql Tool
======================================================

1. Install `command ql <https://godoc.org/modernc.org/ql/ql>`_
2. If the Virtual Device Service runs in a Docker container, it must mount the directory (/db) that contains the ql database in the container. For example::

      device-virtual:
      image: edgexfoundry/docker-device-virtual-go:1.0.0
      ports:
        - "49990:49990"
      container_name: device-virtual
      hostname: device-virtual
      networks:
        - edgex-network
      volumes:
        - db-data:/data/db
        - log-data:/edgex/logs
        - consul-config:/consul/config
        - consul-data:/consul/data
        - /mnt/hgfs/EdgeX/DeviceVirtualDB:/db # Mount ql database directory
      depends_on:
        - data
        - command

3. If the Virtual Device Service runs in dev mode, the ql database directory is under the driver directory

Command examples:

* Query all data::

    $ ql -db /path-to-the-ql-db-folder/deviceVirtual.db -fld "select * from VIRTUAL_RESOURCE"

* Update Enable_Randomization::

    ql -db /path-to-the-ql-db-folder/deviceVirtual.db "update VIRTUAL_RESOURCE set ENABLE_RANDOMIZATION=false where DEVICE_NAME=\"Random-Integer-Generator01\" and DEVICE_RESOURCE_NAME=\"RandomValue_Int8\" "

* Update Value::

    $ ql -db /path-to-the-ql-db-folder/deviceVirtual.db "update VIRTUAL_RESOURCE set VALUE=\"26\" where DEVICE_NAME=\"Random-Integer-Generator01\" and DEVICE_RESOURCE_NAME=\"RandomValue_Int8\" "
