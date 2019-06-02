
Application Functions SDK (GoLang)
==================================

Welcome to the Application Functions SDK for EdgeX. This SDK is meant to provide all the plumbing necessary for developers to get started in processing/transforming/exporting data out of EdgeX.  The Application Functions SDK assists developers create application services.  Application Services will replace EdgeX export services in a future release.

Getting started
---------------

The SDK is built around the idea of a "Functions Pipeline". A functions pipeline is a collection of various functions that process the data in the order that you've specified. The functions pipeline is executed by the specified trigger in the configuration.toml . The first function in the pipeline is called with the event that triggered the pipeline (ex. events.Model). Each successive call in the pipeline is called with the return result of the previous function. Let's take a look at a simple example that creates a pipeline to filter particular device ids and subsequently transform the data to XML:

.. code:: go

  package main

  import (
  	"fmt"
  	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
  	"os"
  )

  func main() {

  	// 1) First thing to do is to create an instance of the EdgeX SDK, giving it a service key
  	edgexSdk := &appsdk.AppFunctionsSDK{
  		ServiceKey: "SimpleFilterXMLApp", // Key used by Rigistry (Aka Consul)
  	}

  	// 2) Next, we need to initialize the SDK
  	if err := edgexSdk.Initialize(); err != nil {
  		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
  		os.Exit(-1)
  	}

  	// 3) Since our DeviceNameFilter Function requires the list of Device Names we would
  	// like to search for, we'll go ahead and define that now.
  	deviceIDs := []string{"GS1-AC-Drive01"}

  	// 4) This is our pipeline configuration, the collection of functions to
  	// execute every time an event is triggered.
  	if err := edgexSdk.SetPipeline(edgexSdk.DeviceNameFilter(deviceIDs), edgexSdk.XMLTransform()); err != nil {
  		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK SetPipeline failed: %v\n", err))
  		os.Exit(-1)
  	}

  	// 5) shows how to access the application's specific configuration settings.
  	appSettings := edgexSdk.ApplicationSettings()
  	if appSettings != nil {
  		appName, ok := appSettings["ApplicationName"]
  		if ok {
  			edgexSdk.LoggingClient.Info(fmt.Sprintf("%s now running...", appName))
  		} else {
  			edgexSdk.LoggingClient.Error("ApplicationName application setting not found")
  			os.Exit(-1)
  		}
  	} else {
  		edgexSdk.LoggingClient.Error("No application settings found")
  		os.Exit(-1)
  	}

  	// 6) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events to trigger the pipeline.
  	edgexSdk.MakeItRun()
  }

The above example is meant to merely demonstrate the structure of your application. Notice that the output of the last function is not available anywhere inside this application. You must provide a function in order to work with the data from the previous function. Let's go ahead and add the following function that prints the output to the console.

.. code:: go

    func printXMLToConsole(edgexcontext *excontext.Context, params ...interface{}) (bool,interface{}) {
      if len(params) < 1 {
      	// We didn't receive a result
      	return false, errors.New("No Data Received")
      }
      println(params[0].(string))
      return true, nil
    }

After placing the above function in your code, the next step is to modify the pipeline to call this function:

.. code:: go

  edgexSdk.SetPipeline(
    edgexSdk.DeviceNameFilter(deviceIDs),
    edgexSdk.XMLTransform(),
    printXMLToConsole //notice this is not a function call, but simply a function pointer.
  )

After making the above modifications, you should now see data printing out to the console in XML when an event is triggered.

*You can find this example in the /examples directory located in this repository. You can also use the provided `EdgeX Applications Function SDK.postman_collection.json" file to load into postman to trigger the sample pipeline.*

Up until this point, the pipeline has been triggered by an event over HTTP and the data at the end of that pipeline lands in the last function specified. In the example, data ends up printed to the console. Perhaps we'd like to send the data back to where it came from. In the case of an HTTP trigger, this would be the HTTP response. In the case of a message bus, this could be a new topic to send the data back to for other applications that wish to receive it. To do this, simply call edgexcontext.Complete([]byte outputData) passing in the data you wish to "respond" with. In the above printXMLToConsole(...) function, replace println(params[0].(string)) with edgexcontext.Complete([]byte(params[0].(string))). You should now see the response in your postman window when testing the pipeline.

Triggers
--------

Triggers determine how the app functions pipeline begins execution. In the simple example provided above, an HTTP trigger is used. The trigger is determine by the configuration.toml file located in the /res directory under a section called [Binding]. Check out the Configuration Section for more information about the toml file.

.. toctree::
   :maxdepth: 1

   Ch-MessageBusTrigger
   Ch-HTTPTrigger

Context API
-----------
The context parameter passed to each function/transform provides operations and data associated with each execution of the pipeline. Let's take a look at a few of the properties that are available:

.. code:: go

  type Context struct {
  	EventID       string // ID of the EdgeX Event -- will be filled for a received JSON Event
  	EventChecksum string // Checksum of the EdgeX Event -- will be filled for a received CBOR Event
  	CorrelationID string // This is the ID used to track the EdgeX event through entire EdgeX framework.
  	Configuration common.ConfigurationStruct // This holds the configuration for your service. This is the preferred way to access your custom application settings that have been set in the configuration.
  	LoggingClient logger.LoggingClient // This is exposed to allow logging following the preferred logging strategy within EdgeX.
  }

**Logging**

The LoggingClient exposed on the context is available to leverage logging libraries/service leveraged throughout the EdgeX framework. The SDK has initialized everything so it can be used to log Trace, Debug, Warn, Info, and Error messages as appopriate. See examples/simple-filter-xml/main.go for an example of how to use the LoggingClient.

**.MarkAsPushed()**

The .MarkAsPushed() function is used to indicate to EdgeX Core Data that an event has been "pushed" and is no longer required to be stored. The scheduler service will purge all events that have been marked as pushed based on the configured schedule. By default, it is once daily at midnight. If you leverage the built in export functions (i.e. HTTP Export, or MQTT Export), then the event will automatically be marked as pushed upon a successful export.

**.Complete()**

The .Complete([]byte outputData) function can be used to return data back to the configured trigger. In the case of an HTTP trigger, this would be an HTTP Response to the caller. In the case of a message bus trigger, this is how data can be published to a new topic per the configuration.

Built-In Transforms/Functions
-----------------------------

**Filtering**

There are two basic types of filtering included in the SDK to add to your pipeline. The provided Filter functions return a type of events.Model.

- DeviceNameFilter([]string deviceNames) - This function will filter the event data down to the specified device names before calling the next function.
- ValueDescriptorFilter([]string valueDescriptors) - This function will filter the event data down to the specified device value descriptor before calling the next function.

**Conversion**

There are two conversions included in the SDK that can be added to your pipeline. These transforms return a string.

- XMLTransform() - This function received an events.Model type and converts it to XML format.
- JSONTransform() - This function received an events.Model type and converts it to JSON format.

**Export Functions**

There are two export functions included in the SDK that can be added to your pipeline.

- HTTPPost(string url, mimeType string) - This function requires an endpoint be passed in order to configure the URL to POST data to as well as the mime type. Currently, only unauthenticated endpoints are supported. Authenticated endpoints will be supported in the future. If will be POSTing JSON or XML you can leverage the HTTPPostJSON(url string) or HTTPPostXML(url string) respectively as shortcuts so you don't have to specify mimeType yourself. This function will mark the received EdgeX event as pushed in Core Data upon a success response code.
- MQTTSend(addr models.Addressable, cert string, key string, qos byte, retain bool, autoreconnect bool) - This function will send data from the previous function in the pipeline to the specified MQTT broker. If no previous function exists, then the event that triggered the pipeline will be used. This function will mark the received EdgeX event as pushed in Core Data upon a success response code.

Configuration
-------------

Similar to other EdgeX services, configuration is first determined by the configuration.toml file in the /res folder. If -r is passed to the application on startup, the SDK will leverage the provided registry (i.e Consul) to push configuration from the file into the registry and monitor configuration from there. There are two primary sections in the configuration.toml file that will need to be set that are specific to the AppFunctionsSDK.

- [Binding] - This specifies the trigger type and associated data required to configurate a trigger.

.. code::

  [Binding]
  Type=""
  SubscribeTopic=""
  PublishTopic=""

- [ApplicationSettings] - Is used for custom application settings and is accessed via the ApplicationSettings() API. The ApplicationSettings API returns a map[string] string containing the contents on the ApplicationSetting section of the configuration.toml file.

.. code ::

  [ApplicationSettings]
  ApplicationName = "My Application Service"

Error Handling
--------------

- Each transform returns a true or false as part of the return signature. This is called the continuePipeline flag and indicates whether the SDK should continue calling successive transforms in the pipeline.
- return false, nil will stop the pipeline and stop processing the event. This is useful for example when filtering on values and nothing matches the criteria you've filtered on.
- return false, error, will stop the pipeline as well and the SDK will log the errorString you have returned.
- Returning true tells the SDK to continue, and will call the next function in the pipeline with your result.
- The SDK will handle exiting when receiving a SIGTERM/SIGINT event.
