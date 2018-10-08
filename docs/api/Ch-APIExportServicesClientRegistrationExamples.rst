##########################################################
APIs - Export Services - Client Registration API Examples
##########################################################

============
Introduction
============

This page of Export Services Client Registration API Examples contains various export service registrations to show how to get data from the EdgeX Foundry export distribution service in the format that is wanted and on the output channel that is wanted (MQTT or REST).  Example Core Data Event/Readings are used to show the expected output by the export clients.

=====
Setup
=====

* Start the EdgeX Microservices via the User Guide ( )

* Start an HTTP Server App to listen for posts on port 8111 of the local box
* Start an MQTT Broker and Topic listener.  The following examples use Cloud MQTT with the listed properties: 
        
  * Address:  m10.cloudmqtt.com
  * port:15421
  * publisher: EdgeXExportPublisher
  * user:hukfgtoh
  * password: uP6hJLYW6Ji4
  * topic:  EdgeXDataTopic

The following Value Descriptors need to be added (posted) to Core Data before running the tests or examples

::

   POST the following JSON bodies to http://localhost:48080/api/v1/valuedescriptor
   {"name":"temperature","min":"-40","max":"140","type":"F","uomLabel":"degree cel","defaultValue":"0","formatting":"%s","labels":["temp","hvac"]}
   {"name":"humidity","min":"0","max":"100","type":"F","uomLabel":"per","defaultValue":"0","formatting":"%s","labels":["humidity","hvac"]}

**Test Message**

Post the following Event/Reading message to Core Data in order to test each of the scenarios below and to get the results as depicted under each example

::

    POST the JSON body to:  http://localhost:48080/api/v1/event
    {"origin":1471806386919,"device":"livingroomthermostat","readings":[{"origin":1471806386919,"name":"temperature","value":"72"}, {"origin":1471806386919,"name":"humidity","value":"58"}]}

**NOTE:  for all POST examples below, inclusion of an origin timestamp is optional.**

**Example #1 - simple tests**

Receive all exported Core Data Event/Readings with valid value descriptors in the reading, uncompressed, unencrypted, and with no filters in place (that is with no device filters in place and no value descriptor filters in place).

**Register for JSON formatted data to be sent to MQTT topic**

::

   POST to http://localhost:48071/api/v1/registration
   {"origin":1471806386919,"name":"MQTTClient","addressable":{"origin":1471806386919,"name":"EdgeXTestMQTTBroker","protocol":"TCP","address":"m10.cloudmqtt.com","port":15421,"publisher":"EdgeXExportPublisher","user":"hukfgtoh", "password":"uP6hJLYW6Ji4","topic":"EdgeXDataTopic"},"format":"JSON","enable":true,"destination":"MQTT_TOPIC"}   

**Result:**

::

  {"pushed":0,"device":"livingroomthermostat","readings":[{"pushed":0,"name":"temperature","value":"72","id":"57ed24f0502fdf73bb637915","created":1475159280744,"modified":1475159280744,"origin":1471806386919},{"pushed":0,"name":"humidity","value":"58","id":"57ed24f0502fdf73bb637916","created":1475159280756,"modified":1475159280756,"origin":1471806386919],"id":"57ed24f0502fdf73bb637917","created":1475159280762,"modified":1475159280762,"origin":1471806386919}


**Register for JSON formatted data to be sent to a REST address**

::

   POST to http://localhost:48071/api/v1/registration
   {"origin":1471806386919,"name":"MQTTXMLClient","addressable":{"origin":1471806386919,"name":"EdgeXTestMQTTXMLBroker","protocol":"TCP","address":"m10.cloudmqtt.com","port":15421,"publisher":"EdgeXExportPublisher", "user":"hukfgtoh","password":"uP6hJLYW6Ji4","topic":"EdgeXXMLDataTopic"},"format":"XML","enable":true,"destination":"REST_ENDPOINT"}

**Result:**

::

   {"origin":1471806386919,"name":"RESTClient","addressable":{"origin":1471806386919,"name":"EdgeXTestREST","protocol":"HTTP","address":"http://localhost","port":8111,"path":"/rest"},"format":"JSON","enable":true}

**Register for XML formatted data to be sent to MQTT topic**

::

   POST to http://localhost:48071/api/v1/registration
   {"origin":1471806386919,"name":"MQTTXMLClient","addressable":{"origin":1471806386919,"name":"EdgeXTestMQTTXMLBroker","protocol":"TCP","address":"m10.cloudmqtt.com","port":15421,"publisher":"EdgeXExportPublisher", "user":"hukfgtoh","password":"uP6hJLYW6Ji4","topic":"EdgeXXMLDataTopic"},"format":"XML","enable":true,"destination":"MQTT_TOPIC"}

**Result:**

::

   <?xml version="1.0" encoding="UTF-8" standalone="yes"?> <Event> <event> <created>1475159280762</created> <id>57ed24f0502fdf73bb637917</id> <modified>1475159280762</modified> <origin>1471806386919</origin> <device>livingroomthermostat</device> <pushed>0</pushed> <readings> <created>1475159280744</created> <id>57ed24f0502fdf73bb637915</id> <modified>1475159280744</modified> <origin>1471806386919</origin> <name>temperature</name> <pushed>0</pushed> <value>72</value> </readings> <readings> <created>1475159280756</created> <id>57ed24f0502fdf73bb637916</id> <modified>1475159280756</modified> <origin>1471806386919</origin> <name>humidity</name> <pushed>0</pushed> <value>58</value> </readings> </event> </Event>

**Register for XML formatted data to be sent to a REST address**

::

   POST to http://localhost:48071/api/v1/registration
   {"origin":1471806386919,"name":"RESTXMLClient","addressable":{"origin":1471806386919,"name":"EdgeXTestRESTXML","protocol":"HTTP","address":"http://localhost","port":8111,"path":"/rest"},"format":"XML","enable":true,"destination":"REST_ENDPOINT"} 


**Result:**

::
  
   <?xml version="1.0" encoding="UTF-8" standalone="yes"?>

    <Event>

        <event>

            <created>1475159280762</created>

            <id>57ed24f0502fdf73bb637917</id>

            <modified>1475159280762</modified>

            <origin>1471806386919</origin>

            <device>livingroomthermostat</device>

            <pushed>0</pushed>

            <readings>

                <created>1475159280744</created>

                <id>57ed24f0502fdf73bb637915</id>

                <modified>1475159280744</modified>

                <origin>1471806386919</origin>

                <name>temperature</name>

                <pushed>0</pushed>

                <value>72</value>

            </readings>

            <readings>

                <created>1475159280756</created>

                <id>57ed24f0502fdf73bb637916</id>

                <modified>1475159280756</modified>

                <origin>1471806386919</origin>

                <name>humidity</name>

                <pushed>0</pushed>

                <value>58</value>

            </readings>

        </event>

    </Event>

**Example #2 - compression tests**

Receive all exported core data Event/Readings with valid value descriptors in the reading, compressed, but unencrypted, and with no filters in place (that is with no device filters in place and no value descriptor filters in place).

**Modify the JSON/MQTT client registration to request future exports be compressed (through GZIP format)**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTClient","compression":"GZIP"}

**Result:**

::

   H8KLCAAAAAAAAADChcKOTQ7CgyAQwoXDrzJrF8KAIsOgVcKaLsKwwow6ScKRBsORwqRpesO3QlcuwrRdw47Du8KZw7fCvcOgwrEuEzrDqFgFDjfCuiF0cMKnwo3DpjHChsOgw5PChMORwoclw5kEFUTCtC7Dqwt0wpfDl8K+N1tfWgnDvQPCo01rw4Qcw57DrH0twqoSw7nCoBwEwqnDkAktwplkYnDCg8Kqw7vCvsKtwpURPMO7wrfDvDrClW/CvFHCksK3wow3TcKjTAU+OBrDqMOACMKRRsKawr8yw5fCrMKtdWvCuHlXR1jDk8Oqw4lResOuwpjCpMO+w4MkTsKYwrQ6YSrDhjHDk8O1w7dSfcKyZMO4w4lSMcKOwpc+InvDuUvDjAEAAA==


**Modify the JSON/REST client registration to request future exports be compressed (through GZIP format)**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTClient","compression":"GZIP"}

**Result:**

::

   H8KLCAAAAAAAAADChcKOTQ7CgyAQwoXDrzJrF8KAIsOgVcKaLsKwwow6ScKRBsORwqRpesO3QlcuwrRdw47Du8KZw7fCvcOgwrEuEzrDqFgFDjfCuiF0cMKnwo3DpjHChsOgw5PChMORwoclw5kEFUTCtC7Dqwt0wpfDl8K+N1tfWgnDvQPCo01rw4Qcw57DrH0twqoSw7nCoBwEwqnDkAktwplkYnDCg8Kqw7vCvsKtwpURPMO7wrfDvDrClW/CvFHCksK3wow3TcKjTAU+OBrDqMOACMKRRsKawr8yw5fCrMKtdWvCuHlXR1jDk8Oqw4lResOuwpjCpMO+w4MkTsKYwrQ6YSrDhjHDk8O1w7dSfcKyZMO4w4lSMcKOwpc+InvDuUvDjAEAAA==


**Modify the XML/MQTT client registration to request future exports be compressed (through GZIP format)**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTXMLClient","compression":"GZIP"}

**Result:**

::

   H8KLCAAAAAAAAADCrcKTTW7CgzAQRsO3OQViw58aw7NnwpDCjMKzak/DkB7CgMOgIVjDgnZkDGpuX8KgwrRuRFFKVcKvw4bDnzwNw4zCk0zCj2/CssO1BjDCncOQwqrDsMOxY8Oge8KgKsONwoU6F8O+w6vDi8OzQ8Ome8KdLRUvW8KtwqDDsMKvw5DDuUd2wqBPAyjDiw7Dnngowrh6wr5XBkoLwpzDocKYJDgNcBzDhznCpsOoM3bCoMOgLCHDgMODLAnCkiDCrHlNwqLDkynCjUgeRhTCicOvwqQcw7/CqBbDq8KZX8K5Q8K1EWfCoSYQZ0EaZWnCjnPCisKWw5RhHAZRAWvDhTDCrmrCtMKWwrYBI8O1wrjCq8KlaGk6w7rDknfDjcO4wpXCgMKiwqVywq1xwqvDiVbDp8KiLQkkw79Bw4IdEcO4VsOEwqbCjGnDtlrDhg4hM8KqSgnDjMKCwrzCgCltb8KAwqI5wrnChcK2XcOMw63CoWx7YCTCpMOowqNywqbDkFrDlQ57GcOZby/DvMKlwr1pw7Y/w5lreinCuMKww5fCv8KrS8KywrvDqihaw54cRcOLQ3wHM8OQQUrDiQMAAA==


**Modify the XML/REST client registration to request future exports be compressed (through GZIP format)**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTXMLClient","compression":"GZIP"}

**Result:**

::

   H8KLCAAAAAAAAADCrcKTTW7CgzAQRsO3OQViw58aw7NnwpDCjMKzak/DkB7CgMOgIVjDgnZkDGpuX8KgwrRuRFFKVcKvw4bDnzwNw4zCk0zCj2/CssO1BjDCncOQwqrDsMOxY8Oge8KgKsONwoU6F8O+w6vDi8OzQ8Ome8KdLRUvW8KtwqDDsMKvw5DDuUd2wqBPAyjDiw7Dnngowrh6wr5XBkoLwpzDocKYJDgNcBzDhznCpsOoM3bCoMOgLCHDgMODLAnCkiDCrHlNwqLDkynCjUgeRhTCicOvwqQcw7/CqBbDq8KZX8K5Q8K1EWfCoSYQZ0EaZWnCjnPCisKWw5RhHAZRAWvDhTDCrmrCtMKWwrYBI8O1wrjCq8KlaGk6w7rDknfDjcO4wpXCgMKiwqVywq1xwqvDiVbDp8KiLQkkw79Bw4IdEcO4VsOEwqbCjGnDtlrDhg4hM8KqSgnDjMKCwrzCgCltb8KAwqI5wrnChcK2XcOMw63CoWx7YCTCpMOowqNywqbDkFrDlQ57GcOZby/DvMKlwr1pw7Y/w5lreinCuMKww5fCv8KrS8KywrvDqihaw54cRcOLQ3wHM8OQQUrDiQMAAA==


**Example #3 - encryption tests**

Receive all exported Core Data Event/Readings with valid value descriptors in the reading, encrypted, but uncompressed, and with no filters in place (that is with no device filters in place and no value descriptor filters in place).

**Modify the JSON/MQTT client registration to request future exports be encrypted (through AES algorithm) and uncompressed (switched back from the last test)**

::

   PUT to http://localhost:48071/api/v1/registration
   {{"name":"MQTTClient","compression":"NONE", "encryption":{"encryptionAlgorithm":"AES","encryptionKey":"123","initializingVector":"123"}}

**Result:**

::

   a8NJLOaJ/IgyPC3PPzJJeA+a4mv5m+OxUkEGh8AQ/wbWRpBdprFV/6iNklEkB2CRFY+PX+X7RoBy0n2+u6aEzUfsdXBTqw0Xel0OGmSSAt95Hv2dNZuZbUOIPqFvsw5GUhoY/cnk7LOyfP8ZVLXilPNKWhsxR72WqZI2buKgkby6XnygDdaUfIYdn/tFhiT/G16T0wGT4ie+9OO1oHzLiy6L1sYhArDsWA8BcWr/bck+3HLQaw3uceyAWPYgWhJEuIxJLWfmfFUrGnxNmDNucYU8McJbEebzA3pZYDshPeD7uxDhayDYaAs9xqQq6oSvhtf4kI7CSaUZ9j8Jhopu3abpVWbSqA5ltUAXQYXp0Nk7cE9grNkGpkjK4CLK3s97tB6xdERIUCSjExBbZRT6Tp73grKQ/||q0XMUiclfK7SEB1kheknfEtOVrq01x+kdUTkKwsCecMjGADv5jYimj5Kev4qiecKMyY1KYYhWAu+ewGZE5IbErPBcCIHAbwakX9nBA14Ym5WWKq7HVAY6y6yTilWEtEr9+bb7xhCFj3mqJk1sAzxL6U+xlPxK5e3+aPRd/iOCd3TJZUw1Ysua3cM1uzXf0JBB+1Cq8mBIUao20= 


**Modify the JSON/REST client registration to request future exports be encrypted (through AES algorithm) and uncompressed (switched back from the last test)**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTClient","compression":"NONE","encryption":{"encryptionAlgorithm":"AES","encryptionKey":"123","initializingVector":"123"}}

**Result:**

::

   a8NJLOaJ/IgyPC3PPzJJeA+a4mv5m+OxUkEGh8AQ/wbWRpBdprFV/6iNklEkB2CRFY+PX+X7RoBy0n2+u6aEzUfsdXBTqw0Xel0OGmSSAt95Hv2dNZuZbUOIPqFvsw5GUhoY/cnk7LOyfP8ZVLXilPNKWhsxR72WqZI2buKgkby6XnygDdaUfIYdn/tFhiT/G16T0wGT4ie+9OO1oHzLiy6L1sYhArDsWA8BcWr/b+3HLQaw3uceyAWPYgWhJEuIxJLWfmfFUrGnxNmDNucYU8McJbEebzA3pZYDshPeD7uxDhayDYaAs9xqQq6oSvhtf4kI7CSaUZ9j8Jhopu3abpVWbSqA5ltUAXQYXp0Nk7cE9grNkGpkjK4CLK3s97tB6xdERIUCSjExBbZRT6Tp73grKQ/q0XMUiclfK7SEB1kheknfEtOVrq01x+kdUTkKwsCecMjGADv5jYimj5Kev4qiecKMyY1KYYhWAu+ewGZE5IbErPBcCIHAbwakX9nBA14Ym5WWKq7HVAY6y6yTilWEtEr9+bb7xhCFj3mqJk1sAzxL6U+xlPxK5e3+aPRd/iOCd3TJZUw1Ysua3cM1uzXf0JBB+1Cq8mBIUao20=


**Modify the XML/MQTT client registration to request future exports be encrypted (through AES algorithm) and uncompressed (switched back from the last test)**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTXMLClient","compression":"NONE","encryption":{"encryptionAlgorithm":"AES","encryptionKey":"123","initializingVector":"123"}}

**Result:**

::

   mR5cLITSvZJOcjc3Zm17SCRe9aKALctrsQlsIZNlGDIh49LsF8AUoDT+P6PTP7OhNCVVo+p4AKDPPv3EWJLNp9yYsnf8R5pEvk8IkuO1LNDiV4w4eZhINPie+85ikHgYld+syjas5o8YxbnESzrTfYTf3SUeqCcj7xTjxPaYc0okXT4ML7F0m7L175OIZPYf0RFMksNgYrJYRYBIRLFlpHOMlU5Ey0K204XFS7lA04mPtIcZp2EIXufWBKI4+tzWz4qGSM3eLX3YJZakDFGWiP5k6C0KswU7GsYEJroJxNwAglKefBcSghqhZg2OoB3IG8YFGpqH7m5cUBnsyTbAIfz4fIa/XRihDMcNl67diXUqOsa1eOkE5A2lC9qX7cd2Yi9t003NfCmvxTGuNgBVNx73jYfB+qSlF9wfC3urqm0Z+D3SvEwYIqty0Hh0lWpDVGWgJReA96lVa/mmukZBcW8PI0JNYrbN5QrsupHvAi9SDGD/gxC7yTzZJQz/sUSMDsUsYjVanUrkl/M0D8caxNVgqUQ+pzA7xMtkq2OSjszWQ0/W+hnklBexz5yWJ3h4+cYIg9P7QA+Ad+7Bv/hXku3TBQSMzmxVNN2VLswk6V+o1eoRYDCiZ5xz17jn9HzcLdfQFEVU7SH2rmv9IM4ka2Xi1ogxKeV52k6v4ilc1PfGfJfRAN4fQn6sR/XfLAo5juWoUAyYHeT07YcyP1XIvAY4ZeVO/0H3nbxMtkE64hJxOGlPMgG2Do1SvPxiSWkvdF3mup8MqzPTW6REOJJqHBj/2jEhEFj53scshpprasSdN6ZWGxgDEIATuiv+J9gKOvnLpmqaZ1iTUrzpBPWyUgpdWUSwJtJV/8tpfX8EAELV1k47Eri0Yh/HGvjr7ySFub16w+AnpH9bWZeK1VOlmWFq/6usN228U57GdmOBJWYLBy0eWU/zjHXj/PcVMaCBgz6Bqq4snXspBEzWG2JkQ/g01RI1qrGa7TsIYyDWThHsPHuLq7uZZtMAsnpxJ6od66+kxYRbfOz05llJUsnPGh+RJ45IA1wIPMKI7zzaGs43eT4IpzmrU2QXhRgPgAzjFIvbjWPQhUXkm2r+bhJgFd+CuFa+aIpTwxCuYYzsncIcA7cZ66VQVLM1EUwOxzKlywJ/owJJ5W9jV/duMA5T2oskiPT7nktlui+b2WnA5H+G+SK7sppQzguY/2nadaKPJD5WUR+UuQSMwIMwopYCKSuC6bmbeEgawWgCscpSw++utt4Z3fkLcCa6LUDPFyYqbC7fu+aeZkpWl4abN7WX/w==

**Modify the XML/REST client registration to request future exports be encrypted (through AES algorithm) and uncompressed (switched back from the last test)**

::

  PUT to http://localhost:48071/api/v1/registration
  {"name":"RESTXMLClient","compression":"NONE","encryption":{"encryptionAlgorithm":"AES","encryptionKey":"123","initializingVector":"123"}}

**Result:**

:: 

   mR5cLITSvZJOcjc3Zm17SCRe9aKALctrsQlsIZNlGDIh49LsF8AUoDT+P6PTP7OhNCVVo+p4AKDPPv3EWJLNp9yYsnf8R5pEvk8IkuO1LNDiV4w4eZhINPie+85ikHgYld+syjas5o8YxbnESzrTfYTf3SUeqCcj7xTjxPaYc0okXT4ML7F0m7L175OIZPYf0RFMksNgYrJYRYBIRLFlpHOMlU5Ey0K204XFS7lA04mPtIcZp2EIXufWBKI4+tzWz4qGSM3eLX3YJZakDFGWiP5k6C0KswU7GsYEJroJxNwAglKefBcSghqhZg2OoB3IG8YFGpqH7m5cUBnsyTbAIfz4fIa/XRihDMcNl67diXUqOsa1eOkE5A2lC9qX7cd2Yi9t003NfCmvxTGuNgBVNx73jYfB+qSlF9wfC3urqm0Z+D3SvEwYIqty0Hh0lWpDVGWgJReA96lVa/mmukZBcW8PI0JNYrbN5QrsupHvAi9SDGD/gxC7yTzZJQz/sUSMDsUsYjVanUrkl/M0D8caxNVgqUQ+pzA7xMtkq2OSjszWQ0/W+hnklBexz5yWJ3h4+cYIg9P7QA+Ad+7Bv/hXku3TBQSMzmxVNN2VLswk6V+o1eoRYDCiZ5xz17jn9HzcLdfQFEVU7SH2rmv9IM4ka2Xi1ogxKeV52k6v4ilc1PfGfJfRAN4fQn6sR/XfLAo5juWoUAyYHeT07YcyP1XIvAY4ZeVO/0H3nbxMtkE64hJxOGlPMgG2Do1SvPxiSWkvdF3mup8MqzPTW6REOJJqHBj/2jEhEFj53scshpprasSdN6ZWGxgDEIATuiv+J9gKOvnLpmqaZ1iTUrzpBPWyUgpdWUSwJtJV/8tpfX8EAELV1k47Eri0Yh/HGvjr7ySFub16w+AnpH9bWZeK1VOlmWFq/6usN228U57GdmOBJWYLBy0eWU/zjHXj/PcVMaCBgz6Bqq4snXspBEzWG2JkQ/g01RI1qrGa7TsIYyDWThHsPHuLq7uZZtMAsnpxJ6od66+kxYRbfOz05llJUsnPGh+RJ45IA1wIPMKI7zzaGs43eT4IpzmrU2QXhRgPgAzjFIvbjWPQhUXkm2r+bhJgFd+CuFa+aIpTwxCuYYzsncIcA7cZ66VQVLM1EUwOxzKlywJ/owJJ5W9jV/duMA5T2oskiPT7nktlui+b2WnA5H+G+SK7sppQzguY/2nadaKPJD5WUR+UuQSMwIMwopYCKSuC6bmbeEgawWgCscpSw++utt4Z3fkLcCa6LUDPFyYqbC7fu+aeZkpWl4abN7WX/w==


**Example #4 - compressed and encrypted tests**

Receive all exported Core Data Event/Readings with valid value descriptors in the reading, encrypted and compressed, but with no filters in place (that is with no device filters in place and no value descriptor filters in place).

Modify the JSON/MQTT client registration to compress the data (this time through ZIP format) and leave the data encrypted (through AES algorithm) per the last tests

::

   PUT to http://localhost:48071/api/v1/registration
   {{"name":"MQTTClient","compression":"ZIP"}

**Result:**

::

   2ig0NpFqhIsvZ3Uk9Yj5+537fNBHOJ1ALW+yH0X9G4hMeGz1vvYY+cZZSdlEPIf0K/ArY9hpbKLOOGbHlfsmIu/TD1HzIpX0neaBT6kJyrxmV/Pd8HNxFhYvjZuqVQ/xB5oedve5TpfotHCGf2qylx1pntgIGxJKXNAVJGafRwNxq6vU7P97D3nVrU7IThBvLHiGN2KbHWr/dJXAPXNWymytTXfOaDT4yT0WDACHWaE8VhVpzkUHjZfQpouRqODexozx1O/goMbSXuIOMJqNnFzy5ms9fQfDJM9bH6hj+l/K7ZCpQrApTH0uA2X/wm30MQn7eGKdrq1eNLyMbAEV2Q==


**Modify the JSON/REST client registration to compress the data (this time through ZIP format) and leave the data encrypted (through AES algorithm) per the last tests**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTClient","compression":"ZIP"}

**Result:**

::
 
   2ig0NpFqhIsvZ3Uk9Yj5+537fNBHOJ1ALW+yH0X9G4hMeGz1vvYY+cZZSdlEPIf0K/ArY9hpbKLOOGbHlfsmIu/TD1HzIpX0neaBT6kJyrxmV/Pd8HNxFhYvjZuqVQ/xB5oedve5TpfotHCGf2qylx1pntgIGxJKXNAVJGafRwNxq6vU7P97D3nVrU7IThBvLHiGN2KbHWr/dJXAPXNWymytTXfOaDT4yT0WDACHWaE8VhVpzkUHjZfQpouRqODexozx1O/goMbSXuIOMJqNnFzy5ms9fQfDJM9bH6hj+l/K7ZCpQrApTH0uA2X/wm30MQn7eGKdrq1eNLyMbAEV2Q==


**Modify the XML/MQTT client registration to compress the data (this time through ZIP format) and leave the data encrypted (through AES algorithm) per the last tests**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTXMLClient","compression":"ZIP"}

**Result:**

::

   pOtrirw5DBSzx6owfmtR2N5T1KwoRvIQCTxg6ozz+ZbI2h5eVRml+NNJ6h8+kvxNk762mL1scjWPcazpWZP9EdfNW5BhMtyNjmzJPMlkcm9tTaT/lB8E6YOpoPXv+nUBSuxGBc1Ng/kkbO6HqwOl0dcyaXNANnkUREJilSG/pyhCmjQz9zXzETmyJX8Ognsj3ZKAEy6JFMkUftywUQ9ltb19DESKZTGiXg2F8afbkSo+52hLP1ngXWuO5ytmOuthepxSmCwK27J+LjwubAW3Me/UM8sNp3tNYAZRfQSJzvFRuV+trfIlnMY9Ha3/V6qmdPOQScBcWrY8p4ev6PKXiHDszg04hmexnFV3zFnJ4Yhd46XHXz8S9P2IHHZy1oy7zM0b/ikC7+rWSURilki/ryBcz9NwTfyatkAF8kddTaSzokxdHHiVfw9Y9u7SgCvKWf26ZSE9PXCyjALtREH9Ic3x4j0PPhcJpyo0ZzFym52MgXCkrJTe1wPVlGFr6ndgmAmy2FawQzu0HDW+mvxddA==


**Modify the XML/REST client registration to compress the data (this time through ZIP format) and leave the data encrypted (through AES algorithm) per the last tests**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTXMLClient","compression":"ZIP"}

**Result:**

::

   pOtrirw5DBSzx6owfmtR2N5T1KwoRvIQCTxg6ozz+ZbI2h5eVRml+NNJ6h8+kvxNk762mL1scjWPcazpWZP9EdfNW5BhMtyNjmzJPMlkcm9tTaT/lB8E6YOpoPXv+nUBSuxGBc1Ng/kkbO6HqwOl0dcyaXNANnkUREJilSG/pyhCmjQz9zXzETmyJX8Ognsj3ZKAEy6JFMkUftywUQ9ltb19DESKZTGiXg2F8afbkSo+52hLP1ngXWuO5ytmOuthepxSmCwK27J+LjwubAW3Me/UM8sNp3tNYAZRfQSJzvFRuV+trfIlnMY9Ha3/V6qmdPOQScBcWrY8p4ev6PKXiHDszg04hmexnFV3zFnJ4Yhd46XHXz8S9P2IHHZy1oy7zM0b/ikC7+rWSURilki/ryBcz9NwTfyatkAF8kddTaSzokxdHHiVfw9Y9u7SgCvKWf26ZSE9PXCyjALtREH9Ic3x4j0PPhcJpyo0ZzFym52MgXCkrJTe1wPVlGFr6ndgmAmy2FawQzu0HDW+mvxddA==


**Example #5 - filter invalid readings tests**

Attempt to receive Event/Readings with readings with values that fall outside the parameters of the associated Value Descriptor.  This should result in no data reaching the client endpoints (MQTT or REST).

**No client registration changes are required for these tests.**

**Test Message**

::

   For this example, POST the following event to core data (note the reading data is outside the value descriptors for temperature and humidity)
   {"origin":1471806386919,"device":"livingroomthermostat","readings":[{"origin":1471806386919,"name":"temperature","value":"-720"}, {"origin":1471806386919,"name":"humidity","value":"-580"}]

**Results**

No data should be seen in the HTTP client or MQTT Broker client.  All data is filtered.

Look for the RejectedEventsServiceActivator: Rejected Event:  Event [... message in Export Distro log

**Example #6 - receive invalid reading tests**

With the value descriptor check turned off, receive all exported core data Event/Readings regardless of the fact that the readings have values that fall outside the value descriptors for temperature and humidity.

**Setup**

* Stop the EdgeX Export Distro Service
* In the application.properties file of the microservice, **change** the valuedescriptor.check=true to **false**
* Restart the EdgeX Export Distro Service

**No client registration changes are required for these tests.**

**Test Message**

::

   For this example, POST the following event to Core Data (note the reading data is outside the value descriptors for temperature and humidity)
   {"origin":1471806386919,"device":"livingroomthermostat","readings":[{"origin":1471806386919,"name":"temperature","value":"-720"}, {"origin":1471806386919,"name":"humidity","value":"-580"}]

**Results:**

2 messages each should appear in the MQTT topic and REST server output as shown in Example #4 results.

**Cleanup**

* Stop the EdgeX Export Distro Service
* In the application.properties file of the microservice, return the valuedescriptor.check=true
* Restart the EdgeX Export Distro Service
* Remove all existing client registrations with the following DELETE calls
  
 * http://localhost:48071/api/v1/registration/name/RESTClient
 * http://localhost:48071/api/v1/registration/name/RESTXMLClient
 * http://localhost:48071/api/v1/registration/name/MQTTClient
 * http://localhost:48071/api/v1/registration/name/MQTTXMLClient 

**Example #7 - device filter out tests**

Receive all exported Core Data Event/Readings with valid value descriptors in the reading, uncompressed, unencrypted, but with a device filter (but no value descriptor filter) in place.  In this example, the device filter will not match any devices on the incoming events.

Register for JSON formatted data to be sent to MQTT topic

::

   POST to http://localhost:48071/api/v1/registration
   {"origin":1471806386919,"name":"MQTTClient","addressable":{"origin":1471806386919,"name":"EdgeXTestMQTTBroker","protocol":"TCP","address":"m10.cloudmqtt.com","port":15421,"publisher":"EdgeXExportPublisher","user":"hukfgtoh","password":"uP6hJLYW6Ji4","topic":"EdgeXDataTopic"},"format":"JSON","filter":{"deviceIdentifiers":["motorrpm"]},"enable":true,"destination":"MQTT_TOPIC"}


**Result:**

No data should be received by the client

**Register for JSON formatted data to be sent to a REST address**

::

   POST to http://localhost:48071/api/v1/registration
   {"origin":1471806386919,"name":"RESTClient","addressable":{"origin":1471806386919,"name":"EdgeXTestREST","protocol":"HTTP","address":"http://localhost","port":8111,"path":"/rest"},"format":"JSON","filter":{"deviceIdentifiers":["motorrpm"]},"enable":true,"destination":"REST_ENDPOINT"} 


**Result:**

No data should be received by the client

**Register for XML formatted data to be sent to MQTT topic**

::

   POST to http://localhost:48071/api/v1/registration
   {"origin":1471806386919,"name":"MQTTXMLClient","addressable":{"origin":1471806386919,"name":"EdgeXTestMQTTXMLBroker","protocol":"TCP","address":"m10.cloudmqtt.com","port":15421,"publisher":"EdgeXExportPublisher","user":"hukfgtoh","password":"uP6hJLYW6Ji4","topic":"EdgeXXMLDataTopic"},"format":"XML","filter":{"deviceIdentifiers":["motorrpm"]},"enable":true,"destination":"MQTT_TOPIC"}


**Result:**

No data should be received by the client

**Register for XML formatted data to be sent to a REST address**

::

   POST to http://localhost:48071/api/v1/registration
   {"origin":1471806386919,"name":"RESTXMLClient","addressable":{"origin":1471806386919,"name":"EdgeXTestRESTXML","protocol":"HTTP","address":"http://localhost","port":8111,"path":"/rest"},"format":"XML","filter":{"deviceIdentifiers":["motorrpm"]},"enable":true,"destination":"REST_ENDPOINT"}


**Result:**

No data should be received by the client

**Example #8 - device filter in tests**

Receive all exported Core Data Event/Readings with valid value descriptors in the reading, uncompressed, unencrypted, but with a device filter (but no value descriptor filter) in place.  In this example, the device filter will match the devices on the incoming events.
Modify the JSON/MQTT client registration to filter on devices that match the incoming Event device

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTClient", "filter":{"deviceIdentifiers":["livingroomthermostat"]}}

**Result:**

::

   {"pushed":0,"device":"livingroomthermostat","readings":[{"pushed":0,"name":"temperature","value":"72","id":"57ee8afbee7a127aa30cd019","created":1475250939667,"modified":1475250939667,"origin":1471806386919},{"pushed":0,"name":"humidity","value":"58","id":"57ee8afbee7a127aa30cd01a","created":1475250939671,"modified":1475250939671,"origin":1471806386919}],"id":"57ee8afbee7a127aa30cd01b","created":1475250939673,"modified":1475250939673,"origin":1471806386919}


**Modify the JSON/REST client registration to filter on devices that match the incoming Event device**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTClient","filter":{"deviceIdentifiers":["livingroomthermostat"]}}

**Result:**

::

   {"pushed":0,"device":"livingroomthermostat","readings":[{"pushed":0,"name":"temperature","value":"72","id":"57ee8afbee7a127aa30cd019","created":1475250939667,"modified":1475250939667,"origin":1471806386919},{"pushed":0,"name":"humidity","value":"58","id":"57ee8afbee7a127aa30cd01a","created":1475250939671,"modified":1475250939671,"origin":1471806386919}],"id":"57ee8afbee7a127aa30cd01b","created":1475250939673,"modified":1475250939673,"origin":1471806386919}


**Modify the XML/MQTT client registration to filter on devices that match the incoming Event device**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTXMLClient","filter":{"deviceIdentifiers":["livingroomthermostat"]}}

**Result:**

::

   <?xml version="1.0" encoding="UTF-8" standalone="yes"?> <Event> <event> <created>1475250939673</created> <id>57ee8afbee7a127aa30cd01b</id> <modified>1475250939673</modified> <origin>1471806386919</origin> <device>livingroomthermostat</device> <pushed>0</pushed> <readings> <created>1475250939667</created> <id>57ee8afbee7a127aa30cd019</id> <modified>1475250939667</modified> <origin>1471806386919</origin> <name>temperature</name> <pushed>0</pushed> <value>72</value> </readings> <readings> <created>1475250939671</created> <id>57ee8afbee7a127aa30cd01a</id> <modified>1475250939671</modified><origin>1471806386919</origin> <name>humidity</name> <pushed>0</pushed> <value>58</value> </readings> </event> </Event> 

**Modify the XML/REST client registration to filter on devices that match the incoming Event device**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTXMLClient","filter":{"deviceIdentifiers":["livingroomthermostat"]}}

**Result:**

::

    <?xml version="1.0" encoding="UTF-8" standalone="yes"?>

    <Event>

        <event>

            <created>1475250939673</created>

            <id>57ee8afbee7a127aa30cd01b</id>

            <modified>1475250939673</modified>

            <origin>1471806386919</origin>

            <device>livingroomthermostat</device>

            <pushed>0</pushed>

            <readings>

                <created>1475250939667</created>

                <id>57ee8afbee7a127aa30cd019</id>

                <modified>1475250939667</modified>

                <origin>1471806386919</origin>

                <name>temperature</name>

                <pushed>0</pushed>

                <value>72</value>

            </readings>

            <readings>

                <created>1475250939671</created>

                <id>57ee8afbee7a127aa30cd01a</id>

                <modified>1475250939671</modified>

                <origin>1471806386919</origin>

                <name>humidity</name>

                <pushed>0</pushed>

                <value>58</value>

            </readings>

        </event>

    </Event>

**Example #9 - value descriptor filter out tests**

Receive all exported Core Data Event/Readings with valid value descriptors in the reading, uncompressed, unencrypted, but with a value descriptor filter (but no device filter) in place.  In this example, the value descriptor filter will not match any readings' values on the incoming events.

**Modify the JSON/MQTT client registration to filter on value descriptor (and remove device filter) that match the incoming Events' value names**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTClient", "filter":{"valueDescriptorIdentifiers":["rpm", "vibration"]}}

**Result:**

No data should be received by the client

**Modify the JSON/REST client registration to filter on value descriptor (and remove device filter) that match the incoming Events' value names**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTClient","filter":{"valueDescriptorIdentifiers":["rpm", "vibration"]}}

**Result:**

No data should be received by the client

**Modify the XML/MQTT client registration to filter on value descriptor (and remove device filter) that match the incoming Events' value names**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTXMLClient","filter":{"valueDescriptorIdentifiers":["rpm", "vibration"]}}

**Result:**

No data should be received by the client

**Modify the XML/REST client registration to filter on value descriptor (and remove device filter) that match the incoming Events' value names**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTXMLClient","filter":{"valueDescriptorIdentifiers":["rpm", "vibration"]}}

**Result:**

No data should be received by the client

**Example #10 - value descriptor filter in tests**

Receive all exported Core Data Event/Readings with valid value descriptors in the reading, uncompressed, unencrypted, but with a value descriptor filter (but no device filter) in place.  In this example, the value descriptor filter will match the value descriptors on the incoming events.
Modify the JSON/MQTT client registration to filter on value descriptors that match the incoming Reading values

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTClient", "filter":{"valueDescriptorIdentifiers":["temperature", "humidity"]}}

**Result:**

::

   {"pushed":0,"device":"livingroomthermostat","readings":[{"pushed":0,"name":"temperature","value":"72","id":"57ee8f00ee7a127aa30cd028","created":1475251968071,"modified":1475251968071,"origin":1471806386919},{"pushed":0,"name":"humidity","value":"58","id":"57ee8f00ee7a127aa30cd029","created":1475251968079,"modified":1475251968079,"origin":1471806386919}],"id":"57ee8f00ee7a127aa30cd02a","created":1475251968081,"modified":1475251968081,"origin":1471806386919}


**Modify the JSON/REST client registration to filter on value descriptors that match the incoming Reading values**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTClient","filter":{"valueDescriptorIdentifiers":["temperature", "humidity"]}}

**Result:**

::

   {"pushed":0,"device":"livingroomthermostat","readings":[{"pushed":0,"name":"temperature","value":"72","id":"57ee8f13ee7a127aa30cd02b","created":1475251987037,"modified":1475251987037,"origin":1471806386919},{"pushed":0,"name":"humidity","value":"58","id":"57ee8f13ee7a127aa30cd02c","created":1475251987044,"modified":1475251987044,"origin":1471806386919}],"id":"57ee8f13ee7a127aa30cd02d","created":1475251987065,"modified":1475251987065,"origin":1471806386919}


**Modify the XML/MQTT client registration to filter on value descriptors that match the incoming Reading values**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"MQTTXMLClient","filter":{"valueDescriptorIdentifiers":["temperature", "humidity"]}}

**Result:**

::

   <?xml version="1.0" encoding="UTF-8" standalone="yes"?> <Event> <event> <created>1475251968081</created> <id>57ee8f00ee7a127aa30cd02a</id> <modified>1475251968081</modified> <origin>1471806386919</origin> <device>livingroomthermostat</device> <pushed>0</pushed> <readings> <created>1475251968071</created> <id>57ee8f00ee7a127aa30cd028</id> <modified>1475251968071</modified> <origin>1471806386919</origin> <name>temperature</name> <pushed>0</pushed> <value>72</value> </readings> <readings> <created>1475251968079</created> <id>57ee8f00ee7a127aa30cd029</id> <modified>1475251968079</modified><origin>1471806386919</origin> <name>humidity</name> <pushed>0</pushed> <value>58</value> </readings> </event> </Event>
 
**Modify the XML/REST client registration to filter on value descriptors that match the incoming Reading values**

::

   PUT to http://localhost:48071/api/v1/registration
   {"name":"RESTXMLClient","filter":{"valueDescriptorIdentifiers":["temperature", "humidity"]}}

**Result:**

::

    <?xml version="1.0" encoding="UTF-8" standalone="yes"?>

    <Event>

        <event>

            <created>1475251987065</created>

            <id>57ee8f13ee7a127aa30cd02d</id>

            <modified>1475251987065</modified>

            <origin>1471806386919</origin>

            <device>livingroomthermostat</device>

            <pushed>0</pushed>

            <readings>

                <created>1475251987037</created>

                <id>57ee8f13ee7a127aa30cd02b</id>

                <modified>1475251987037</modified>

                <origin>1471806386919</origin>

                <name>temperature</name>

                <pushed>0</pushed>

                <value>72</value>

            </readings>

            <readings>

                <created>1475251987044</created>

                <id>57ee8f13ee7a127aa30cd02c</id>

                <modified>1475251987044</modified>

                <origin>1471806386919</origin>

                <name>humidity</name>

                <pushed>0</pushed>

                <value>58</value>

            </readings>

        </event>

    </Event>



