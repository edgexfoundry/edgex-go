##################################
Writing a Device Service for EdgeX
##################################

The EdgeX Device Service SDK helps developers quickly create new device connectors for EdgeX because it provides the common scaffolding that each Device Service needs to have.  The scaffolding provides a pattern for provisioning devices.  It provides common template code to receive and react to command (a.k.a. actuation) requests.  Finally, the scaffolding provides the common code to help get the data coming from the sensor into EdgeX Core Data (often referred to as data ingestion).  With the SDK, developers are left to focus on the code that is specific to the communications with the device via the protocol of the device.

In these guides, you will create a simple device service that generates a random number in place of getting data from an actual sensor.  In this way, you get to explore some of the scaffolding and work necessary to complete a device service without actually having a device to talk to.

.. toctree::
   :maxdepth: 1

   Ch-GettingStartedSDK-Go
   Ch-GettingStartedSDK-C
