##########
Scheduling
##########

.. image:: EdgeX_SupportingServicesScheduling.png

============
Introduction
============

The Scheduling microservice includes the Scrubber microservice which cleans up the event and reading data (for Core Data) that has already been exported to the gateway.  Optionally, the Scrubber microservice can also be configured to remove stale event/reading data that has not been exported. The removing of stale event/reading data enables the gateway to continue to operate with a static amount of storage, while the system continues to collect new data from devices and sensors in cases where the export facility is not operational or not operating fast enough to keep up with data collection.

The removal of both exported records and stale records occurs on a configurable schedule. By default, Scrubber cleans up the data every 30 minutes.

The Scrubber microservice does not directly remove the data from EdgeX Foundry's persistent storage itself, rather it calls on Core Data to remove the records. Core Data serves as the single point of access to the persistent event/reading data. The Scrubber microservice is an independent service without any clients. That is, there is no API to call Scrubber. Scrubber operates on time triggers.

===============
Data Dictionary
===============

+---------------------+--------------------------------------------------------------------------------------------+
|   **Class Name**    |   **Descrption**                                                                           | 
+=====================+============================================================================================+
| Schedule            | An object defining a timer or alarm.                                                       | 
+---------------------+--------------------------------------------------------------------------------------------+
| ScheduleEvent       | The action taken by a Service when the Schedule fires.                                     | 
+---------------------+--------------------------------------------------------------------------------------------+

