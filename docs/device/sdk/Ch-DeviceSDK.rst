##########
Device SDK
##########

.. image:: EdgeX_DeviceServiceSDKFlowDiagram.png

Key to diagram

+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------+
|  **Colour of Box**  |   **Descrption**                                                                                                                          | 
+=====================+===========================================================================================================================================+
| Orange              | Everything is part of a Base Service.                                                                                                     | 
+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------+
| Light Green         | Initialization.  Gets its own configuration and registers itself.                                                                         | 
+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------+
| Yellow              | Update Controller. receives, processes, and publishes the update.                                                                         | 
+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------+
| Dark Blue           | UInitializing and setting up of schedules.                                                                                                | 
+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------+
| Gray                | Scaffolding code to be receivers into Device Service.  Processes commands.                                                                | 
+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------+
| Purple              | **Initializes itself.**  Set up in metadata. Registers its Device Service discovery process and registration, and sets up Device Services.| 
|                     |                                                                                                                                           |
|                     | **Gets Device Watchers.**  When a Device Service first comes up it has its initial set of devices.  The Device Watcher waits to receive   |
|                     | information that a new device has occurred.!Then a Device Watcher sends metadata messages out about the new device.                       |
+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------+
| Dark Green          | Send data to Core Data.  How to communicate with the devices, and on what schedule, and to receive information back from the devices.     | 
+---------------------+-------------------------------------------------------------------------------------------------------------------------------------------+
