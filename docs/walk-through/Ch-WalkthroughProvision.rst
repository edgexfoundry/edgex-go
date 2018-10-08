####################################
Provision a Device
####################################

In the last act of setup, a Device Service often discovers and provisions new Devices it finds and is going to manage on the part of EdgeX.  Note the word "often" in the last sentence.  Not all Device Services will discover new Devices or provision them right away.  Depending on the type of Device and how the Devices communicate, it is up to the Device Service to determine how/when to provision a Device.  In some rare cases, the provisioning may be triggered by a human request of the Device Service once everything is in place and once the human can provide the information the Device Service needs to physically connect to the Device.  

Adding your device
------------------

.. _`APIs Core Services Metadata`: https://github.com/edgexfoundry/edgex-go/blob/master/api/raml/core-metadata.raml

See `APIs Core Services Metadata`_ 

For the sake of this demonstration, the call to Core Metadata below will provision the human/dog counting monitor camera as if the Device Service discovered it (by some unknown means) and provisioned the Device as part of some startup process.  To create a Device, it must be associated to a `Device Profile <Ch-WalkthroughDeviceProfile.html>`_ (by name or id), a `Device Service <Ch-WalkthroughDeviceService.html>`_ (by name or id), and `Addressable <Ch-WalkthroughData.html#addressables>`_ (by name or id).  When calling each of the POST calls above, the ID was returned by the associated micro service and used in the call below.  In this example, the names of Device Profile, Device Service, and Addressable are used.

::

   POST to http://localhost:48081/api/v1/device

::

   BODY:  {"name":"countcamera1","description":"human and dog counting camera #1","adminState":"unlocked","operatingState":"enabled","addressable":{"name":"camera1 address"},"labels":
   ["camera","counter"],"location":"","service":{"name":"camera control device service"},"profile":{"name":"camera monitor profile"}}

Note that ``camera monitor profile`` was created by the :download:`CameraMonitorProfile.yml <EdgeX_CameraMonitorProfile.yml>` you uploaded in a previous step.

Test the Setup
--------------

With the Device Service and Device now appropriately setup/provisioned in EdgeX, let's try a few of the micro service APIs out to confirm that things have been configured correctly.

Check the Device Service
^^^^^^^^^^^^^^^^^^^^^^^^

See `APIs Core Services Metadata`_

To begin, check out that the Device Service is available via Core Metadata.

::

   GET to http://localhost:48081/api/v1/deviceservice

Note that the associated Addressable is returned with the Device Service.  There are many additional APIs on Core Metadata to retrieve a Device Service.  As an example, here is one to find all Device Services by label - in this case using the label that was associated to the camera control device service.

::

   GET to http://localhost:48081/api/v1/deviceservice/label/camera

Check the Device
^^^^^^^^^^^^^^^^

See `APIs Core Services Metadata`_

Ensure the monitor camera is among the devices known to Core Metadata.

::

   GET to http://localhost:48081/api/v1/device

Note that the associated Device Profile, Device Service and Addressable is returned with the Device.  Again, there are many additional APIs on Core Metadata to retrieve a Device.  As an example, here is one to find all Devices associated to a given Device Profile - in this case using the camera monitor profile Device Profile name.

::

   GET to http://localhost:48081/api/v1/device/profilename/camera+monitor+profile

Next you can start `Calling commands 〉 <Ch-WalkthroughCommands.html>`_

