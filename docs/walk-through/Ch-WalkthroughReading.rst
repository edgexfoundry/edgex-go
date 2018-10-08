####################################
Sending events and reading data
####################################

In the real world, the human/dog counting camera would start to take pictures, count beings, and send that data to EdgeX.  To simulate this activity. in this section, you will make Core Data API calls as if you were the camera's Device and Device Service.

Send an Event/Reading
---------------------

.. _`APIs Core Services Core Data`: https://github.com/edgexfoundry/edgex-go/blob/master/api/raml/core-data.raml

See `APIs Core Services Core Data`_

Data is submitted to Core Data as an Event.  An Event is a collection of sensor readings from a Device (associated to a Device by its ID or name) at a particular point in time.  A Reading in an Event is a particular value sensed by the Device and associated to a `Value Descriptor <Ch-WalkthroughData.html#value-descriptors>`_ (by name) to provide context to the reading.  So, the human/dog counting camera might determine that there are current 5 people and 3 dogs in the space it is monitoring.  In the EdgeX vernacular, the Device Service upon receiving these sensed values from the Device would create an Event with two Readings - one Reading would contain the key/value pair of humancount:5 and the other Reading would contain the key/value pair of caninecount:3.

The Device Service, on creating the Event and associated Reading objects would transmit this information to Core Data via REST call.

::

   POST to http://localhost:48080/api/v1/event

::

   BODY: {"device":"countcamera1","readings":[{"name":"humancount","value":"5"},{"name":"caninecount","value":"3"}]}

If desired, the Device Service can also supply an origin property (see below) to the Event or Reading to suggest the time (in Epoch timestamp/milliseconds format) at which the data was sensed/collected.  If an origin is not provided, no origin will be set for the Event or Reading, however every Event and Reading is provided a Created and Modified timestamp in the database to give the data some time context.

::

   BODY: {"device":"countcamera1","origin":1471806386919, "readings":[{"name":"humancount","value":"1","origin":1471806386919},{"name":"caninecount","value":"0","origin":1471806386919}]}

**Origin Timestamp Recommendation!**

Note:  Smart devices will often timestamp sensor data and this timestamp can be used as the origin timestamp.  In cases where the sensor/device is unable to provide a timestamp ("dumb" or brownfield sensors), it is recommended that the Device Service create a timestamp for the sensor data that is applied as the origin timestamp for the Device.

Reading data
------------

Now that an Event (or two) and associated Readings have been sent to Core Data, you can use the Core Data API to explore that data that is now stored in MongoDB.

Recall from the Test Setup section, you checked that no data was yet stored in Core Data.  Make the same call and this time, 2 Event records should be the count returned.

::

   GET to http://localhost:48080/api/v1/event/count

Retrieve 10 of the Events associated to the countcamera1 Device.

::

   GET to http://localhost:48080/api/v1/event/device/countcamera1/10

Retrieve 10 of the human count Readings associated to the countcamera1 Device (i.e. - get Readings by Value Descriptor)

::

   GET to http://localhost:48080/api/v1/reading/name/humancount/10

As the final step, you will be `Exporting your device data 〉 <Ch-WalkthroughExporting.html>`_

