HTTP Trigger
============

Designating an HTTP trigger will allow the pipeline to be triggered by a RESTful POST call to http://[host]:[port]/trigger/. The body of the POST must be an EdgeX event.

edgexcontext.Complete([]byte outputData) - Will send the specified data as the response to the request that originally triggered the HTTP Request.

In the main() function, note the call to HTTPPostXML or HTTPPostJSON at the end of the pipeline to return the response.

from `Simple Filter XML Post <https://github.com/edgexfoundry/app-functions-sdk-go/tree/master/examples/simple-filter-xml-post>`_

.. code::

  edgexSdk.SetFunctionsPipeline(
    edgexSdk.DeviceNameFilter(deviceNames),
    edgexSdk.XMLTransform(),
    edgexSdk.HTTPPostXML("<Your endpoint goes here>"),
  )
