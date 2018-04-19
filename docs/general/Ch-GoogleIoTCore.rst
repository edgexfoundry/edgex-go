###############
Google IoT Core
###############

Google IoT Core is a GCP service allowing IoT devices to send telemetry to and receive configurations or commands from the cloud, leveraging Google's infrastructure for reliability, scalability, security and integration with other GCP services.  These instructions cover sending telemetry to the cloud only.
Set up a project on IoT Core

1) Sign up for Google IoT Core at https://cloud.google.com/iot-core/

2) Follow the instructions provided by Google to create a project, a registry and at least one device in that registry in IoT Core, and create the appropriate keys for that device.  The instructions are detailed in how-to guides at https://cloud.google.com/iot/docs/device_manager_guide

3) Take note of the following:

* **projectId** for example my-awesome-project
* **projectLocation** for example us-central1
* **registryId** for example temperature
* **deviceId** for example dht11
* **privateKeyFile** use rsa_private_pkcs8
* **encryptionType** use RS256 or ES256

4) Save the **rsa_private_pkcs8** file in export-distro/src/main/resources

5) Verify that **export-distro/src/main/resources/applications.properties** end with the correct privateKeyFile and encryptionType

#-----------------IoT Core Config

outbound.iotcore.privatekeyfile=**rsa_private_pkcs8**
outbound.iotcore.algorithm=**RS256**

Set up and register IoT Core as an Export Services client

.. _`Client API Registration Examples`: Ch-APIExportServicesClientRegistrationExamples.html
..

1) Follow the instructions under Setup only at `Client API Registration Examples`_.

* start services as indicated
* register a value descriptor:

::
   
   curl -X POST http://<gateway>:48080/api/v1/valuedescriptor --header "Content-Type:application/json" --data '{"name":"temperature","min":"-40","max":"140","type":"F","uomLabel":"degree cel","defaultValue":"0","formatting":"%s","labels":["temp","hvac"]}'

::

   curl -X POST http://<gateway>:48080/api/v1/valuedescriptor --header "Content-Type:application/json" --data '{"name":"humidity","min":"0","max":"100","type":"F","uomLabel":"per","defaultValue":"0","formatting":"%s","labels":["humidity","hvac"]}'

2) Register for JSON formatted data to be sent to IoT Core

(replace the labels with the actual values for your project)

::

    cat <<EOF | curl -X POST http://<gateway>:48071/api/v1/registration --header "Content-Type:application/json" --data @-

    {

      "origin": 1471806386919,

      "name": "IotCoreMQTTClient",

      "addressable": {

        "origin": 1471806386919,

        "name": "IotCoreMQTTClient",

        "protocol": "TCP",

        "address": "",

        "port": 0,

        "publisher": "projects/projectId/locations/projectLocation/registries/registryId/devices/deviceId",

        "user": "unused",

        "password": "",

        "topic": "/devices/deviceId/events"

      },

      "format": "IOTCORE_JSON",

      "destination": "IOTCORE_MQTT",

      "enable": true

    }

    EOF

Note that the string “projects/projectId/locations/projectLocation/registries/registryId/devices/deviceId" is the clientId as defined in IoT Core and must follow this exact format.   You can leave the topic "/devices/deviceId/events" empty as it is automatically generated from the clientId using the format mandated by IoT Core, however if you prefer to use subtopics for your messages, here is the place to set them (for example "/devices/deviceId/events/alarms").
Test it

1) Send some data!

::

   curl -X POST http://<gateway>:48080/api/v1/event --header "Content-Type:application/json" --data '{"origin":1471806386919,"device":"livingroomthermostat","readings":[{"origin":1471806386919,"name":"temperature","value":"72"}, {"origin":1471806386919,"name":"humidity","value":"58"}]}'

2) Verify that the data arrived at IoT Core

IoT Core forwards incoming data automatically to PubSub queues.  While learning how to use PubSub is beyond the scope of this how-to,

it is possible to verify if data arrived by logging in the GCP console and checking the Device details screen under the Registry you created for this project.

The Last message publish time field on that screen should give an indication of the success of the last step.





