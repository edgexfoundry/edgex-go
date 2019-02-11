##################################
Writing a Device Service in C
##################################

In this guide, you create a simple device service in C that generates a random number in place of getting data from an actual sensor.  In this way, you get to explore some of the scaffolding and work necessary to complete a device service without actually having a device to talk to.

====================
Install dependencies
====================

To build a device service using the EdgeX C SDK you'll need the following:
 * libmicrohttpd
 * libcurl
 * libyaml

You can install these on Ubuntu by running::

    sudo apt install libcurl4 libmicrohttpd12 libyaml-0-2

===============================
Get the EdgeX Device SDK for C
===============================

The next step is to download and build the EdgeX Device SDK for C. You always want to use the release of the SDK that matches the release of EdgeX you are targeting. As of this writing the `delhi` release is the current stable release of EdgeX, so we will be using the `delhi` branch of the C SDK.

#. First, clone the delhi branch of device-sdk-c from Github::

    git clone -b delhi https://github.com/edgexfoundry/device-sdk-c.git
    cd ./device-sdk-c

#. Then, build the device-sdk-c::

    ./scripts/build.sh

    

=====================================
Starting a new Device Service project
=====================================

For this guide we're going to use the example template provided by the C SDK as a starting point, and will modify it to generate random integer values. 

#. Begin by copying the contents of src/c/examples/ into a new directory named `example-device-c`::

    cp -r ./src/c/examples ../example-device-c
    cd ../example-device-c

#. You can delete the CMakeLists.txt as we don't need it anymore::

    rm CMakeLists.txt


=========================
Build your Device Service
=========================

Now you are ready to build your new device service using the C SDK you compiled in an earlier step.

#. Tell the compiler where to find the C SDK files::

    export CSDK_DIR=../device-sdk-c/build/release/_CPack_Packages/Linux/TGZ/csdk-0.7.2

.. note::  The exact path to your compiled CSDK_DIR may differ, depending on the tagged version number on the SDK

#. Now you can build your device service executable::

    make 

=============================
Customize your Device Service
=============================

Up to now you've been building the example device service provided by the C SDK. In order to change it to a device service that generates random numbers, you need to modify your `template.c` method **template_get_handler** so that the **while** block looks like this:

.. code-block:: c
    :linenos:
    :lineno-start: 92
    :emphasize-lines: 3,8,11,14

    while (current!=NULL)
    {
      if (strcmp (current->name, "type") ==0 )
      {
        /* Set the resulting reading type as Uint64 */
        readings[i].type = Uint64;

        if (strcmp (current->value, "randdom") ==0 )
        {
          /* Set the reading as a random value between 0 and 100 */
          readings[i].value.ui64_result = rand() % 100;
        }
      }
      current = current->next;
    }


============================
Creating your Device Profile
============================

A Device Profile is a YAML file that describes a class of device to EdgeX.  General characteristics about the type of device, the data these devices provide, and how to command the device is all provided in a Device Profile.  Device Services use the Device Profile to understand what data is being collected from the Device (in some cases providing information used by the Device Service to know how to communicate with the device and get the desired sensor readings).  A Device Profile is needed to describe the data that will be collected from the simple random number generating Device Service.

#. Explore the files in the src/c/examples/res folder.  Take note of the example Device Profile YAML file that is already there (ExampleProfile.yml).  You can explore the contents of this file to see how devices are represented by YAML.  In particular, note how fields or properties of a sensor are represented by “deviceResources”.  Commands to be issued to the device are represented by “commands”.

#. Download this :download:`random-generator-device.yaml <random-generator-device.yaml>` into the ./res folder.  

You can open random-generator-device.yaml in a text editor.  In this Device Profile, you are suggesting that the device you are describing to EdgeX has a single property (or deviceResource) which EdgeX should know about - in this case, the property is the “randomnumber”.  Note how the deviceResource is typed.

    In more real world IoT situations, this deviceResource list could be extensive and could be filled with all different types of data.

    Note also how the Device Profile describes REST commands that can be used by others to call on (or “get”) the random number from the Device Service.   

===============================
Configuring your Device Service
===============================

You will now update the configuration for your new Device Service – changing the port it operates on (so as not to conflict with other Device Services), altering the scheduled times of when the data is collected from the Device Service (every 10 seconds), and setting up the initial provisioning of the random number generating device when the service starts.

* Downlod this :download:`configuration.toml <configuration.toml>` to the ./res folder (this will overwrite an existing file – that’s ok).  

If you will be running EdgeX inside of Docker containers (which you will at the bottom of this guide) you need to tell your new Device Service to listen on the Docker host IP address (172.17.0.1) instead of **localhost**. To do that, modify the configuration.toml file so that the top section looks like this:

.. code-block:: ini
    :linenos:
    :emphasize-lines: 2

    [Service]
    Host = "172.17.0.1"
    Port = 49992


===========================
Rebuild your Device Service
===========================

Now you have your new Device Service, modified to return a random number, a Device Profile that will tell EdgeX how to read that random number, as well as a configuration file that will let your Device Service register itself and it's Device Profile with EdgeX, and begin taking readings every 10 seconds.

#. Rebuild your Device Service to reflect the changes that you have made::

    make 


=======================
Run your Device Service
=======================

Allow your newly created Device Service, which was formed out of the Device Service C SDK, to create sensor mimicking data which it then sends to EdgeX.

#. Follow the :doc:`./Ch-GettingStartedUsers` guide to start all of the EdgeX services in Docker.  From the folder containing the docker-compose file, start EdgeX with a call to::

    docker-compose up -d

#. Back in your custom Device Service directory, tell your device service where to find the `libcsdk.so`::

    export LD_LIBRARY_PATH=$CSDK_DIR/lib

#. Run your device service::

    ./device-example-c

#. You should now see your Device Service having it's /Random command called every 10 seconds. You can verify that it is sending data into EdgeX by watching the logs of the `edgex-core-data` service::

    docker logs -f edgex-core-data

Which would print an Event record every time your Device Service is called. Note that the value of the "randomnumber" reading is an integer between 0 and 100::

    INFO: 2019/02/05 20:27:05 Posting Event: {"id":"","pushed":0,"device":"RandNum-Device01","created":0,"modified":0,"origin":1549398425000,"schedule":null,"event":null,"readings":[{"id":"","pushed":0,"created":0,"origin":0,"modified":0,"device":null,"name":"randomnumber","value":"63"}]}
    INFO: 2019/02/05 20:27:05 Putting event on message queue
