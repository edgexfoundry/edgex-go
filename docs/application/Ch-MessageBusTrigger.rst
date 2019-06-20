Message Bus Trigger
===================

A message bus trigger will execute the pipeline every time data is received off of the configured topic.

Type and Topic Configuration
----------------------------

Here's an example:

.. code::

  Type="messagebus"
  SubscribeTopic="events"
  PublishTopic=""

The Type= is set to "messagebus". EdgeX Core Data is publishing data to the events topic. So to receive data from core data, you can set your SubscribeTopic= either to "" or "events". You may also designate a PublishTopic= if you wish to publish data back to the message bus. edgexcontext.Complete([]byte outputData) - Will send data back to back to the message bus with the topic specified in the PublishTopic= property

Message bus connection configuration
------------------------------------

The other piece of configuration required are the connection settings:

.. code::

  [MessageBus]
  Type = 'zero' #specifies of message bus (i.e zero for ZMQ)
      [MessageBus.PublishHost]
          Host = '*'
          Port = 5564
          Protocol = 'tcp'
      [MessageBus.SubscribeHost]
          Host = 'localhost'
          Port = 5563
          Protocol = 'tcp'

By default, EdgeX Core Data publishes data to the events topic on port 5563. The publish host is used if publishing data back to the message bus.

**Important Note:** Publish Host MUST be different for every topic you wish to publish to since the SDK will bind to the specific port. 5563 for example cannot be used to publish since EdgeX Core Data has bound to that port. Similarly, you cannot have two separate instances of the app functions SDK running publishing to the same port.

In the main() function, note the call to MQTTSend at the end of the pipeline to return the response.

from `Simple Filter XML MQTT <https://github.com/edgexfoundry/app-functions-sdk-go/tree/master/examples/simple-filter-xml-mqtt>`_

.. code::

  edgexSdk.SetFunctionsPipeline(
    edgexSdk.DeviceNameFilter(deviceNames),
    edgexSdk.XMLTransform(),
    printXMLToConsole,
    edgexSdk.MQTTSend(addressable, "", "", 0, false, false),
  )
