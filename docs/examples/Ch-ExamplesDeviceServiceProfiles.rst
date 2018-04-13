###############################
Device Service Profile Examples
###############################

.. _`BAC-121036CE`: http://www.kmccontrols.com.hk/products/productdetail1e33-2.html?partid=BAC-121063C
..

.. _`KMC for sale`: http://www.controlco.com/Manufacturers/KMC-Controls-BACnet-Controllers/BAC-121063CE
..

.. _`BAC-121036CE repo`: https://github.com/edgexfoundry/device-bacnet/blob/master/src/main/resources/KMC.BAC-121036CE.profile.yaml
..



.. _`CC2650STK`: http://www.ti.com/tool/TIDC-CC2650STK-SENSORTAG
..

.. _`TI for sale`: https://store.ti.com/cc2650stk.aspx
..

.. _`CC2650STK repo`: https://github.com/edgexfoundry/device-bluetooth/blob/master/deviceprofile_samples/TI_CC2650_profile.yml
..


.. _`XDK`: https://xdk.bosch-connectivity.com/
..

.. _`Bosch for sale`: https://eu.mouser.com/new/bosch-connected-devices/bosch-xdk110/
..

.. _`XDK repo`: https://github.com/edgexfoundry/device-bluetooth/blob/master/deviceprofile_samples/Bosch_XDK_profile.yaml
..



.. _`CC2540TDK-LIGHT`: http://www.ti.com/tool/CC2540TDK-LIGHT 
..

.. _`TI light for sale`: http://www.ti.com/tool/CC2540TDK-LIGHT
..

.. _`CC2540TDK repo`: https://github.com/edgexfoundry/device-bluetooth/blob/master/deviceprofile_samples/TI_CC2540TDK_profile.yml
..


.. _`Punching Machine`: https://www.fischertechnik.de/en/products/simulating/training-models/96785-sim-punching-machine-with-conveyor-belt-24v-simulation
..

.. _`FT for sale`: https://www.fischertechnik.de/en/products/teaching/training-models/96785-edu-punching-machine-with-conveyor-belt-24v-education
..

.. _`FT repo`: https://github.com/edgexfoundry/device-fischertechnik/blob/master/src/main/resources/Fischertechnik_punching_profile.yml
..


.. _`PS3037`: https://shop.dentinstruments.com/products/powerscout-3037-ps3037
..

.. _`Dent for sale`: https://www.powermeterstore.com/product/dent-powerscout-ps3037-s-n-revenue-grade-networked-power-meter?gclid=CjwKCAiApJnRBRBlEiwAPTgmxFZAN7OvaoISGbzEjWf5mBBe6KYocTXmIswQm1us5GE5ZvJXadtcOBoCkWYQAvD_BwE
..

.. _`PS3037 repo`: https://github.com/edgexfoundry/device-modbus/blob/master/src/main/resources/DENT.Mod.PS6037.profile.yaml
..

.. _`DL05 PLC`: https://www.automationdirect.com/adc/Overview/Catalog/Programmable_Controllers/DirectLogic_Series_PLCs_(Micro_to_Small,_Brick_-a-_Modular)/DirectLogic_05_(Micro_Brick_PLC)/PLC_Units
..

.. _`DL05 for sale`: https://www.automationdirect.com/adc/Shopping/Catalog/Programmable_Controllers/DirectLogic_Series_PLCs_(Micro_to_Small,_Brick_-a-_Modular)/DirectLogic_05_(Micro_Brick_PLC)/PLC_Units
..

.. _`DL05 repo`: https://github.com/edgexfoundry/device-modbus/blob/master/src/main/resources/DL-05.profile.yaml
..


.. _`GS1-10P5`: https://cdn.automationdirect.com/static/manuals/gs1m/gs1m.pdf
..

.. _`AD for sale`: https://www.automationdirect.com/adc/Shopping/Catalog/Drives/GS1_(120_-z-_230_VAC_V-z-Hz_Control)/GS1_Drive_Units_(120_-z-_230_VAC)/GS1-10P5
..

.. _`GS1-10P5 repo`: https://github.com/edgexfoundry/device-modbus/blob/master/src/main/resources/GS1-10P5.profile.yaml
..


.. _`MBUS_RTH_LCD`: http://www.datanab.com/sensors/modbus_rth_lcd.php
..

.. _`Datanab for sale`: http://www.datanab.com/sensors/modbus_rth_lcd.php
..

.. _`MBUS_RTH_LCD repo`: https://github.com/edgexfoundry/device-modbus/blob/master/src/main/resources/MBUS-RTH-LCD.profile.yaml
..


.. _`NHL-3FB1`: http://www.patlite.com/product/detail0000000224.html
..

.. _`Patlite for sale`: https://automationdistribution.com/nhl-3fb1u-ryg/
..

.. _`NHL-3FB1 repo`: https://github.com/edgexfoundry/device-snmp/blob/master/src/main/resources/patlite.NHL-FBL.profile.yaml
..

.. _`iologik E2210`: https://www.moxa.com/product/ioLogik-E2210.htm
..

.. _`Moxa for sale`: https://store.moxa.com/a/product/iologik-e2210-series?id=M20090324001
..

.. _`iologik E2210 repo`: https://github.com/edgexfoundry/device-snmp/blob/master/src/main/resources/moxa.e2210.profile.yaml
..



+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Device             | Device Model      | Where to purchase  | Device Capabilities                   |  Supported Device |  Device Profile         | Notes               | Sample Adressable              |
| Manufacturer       |                   |                    |                                       |  Services         |  Location               |                     |                                |
+====================+===================+====================+=======================================+===================+=========================+=====================+================================+
| KMC                | `BAC-121036CE`_   | `KMC for sale`_    | BACnet Thermostat, 6 Relay, 3 Analog  | device-bacnet     | `BAC-121036CE repo`_    |                     | TCP:{"name":"KMC-121036CE",    |
|                    |                   |                    |                                       |                   |                         |                     | "protocol":"HTTP",             |
|                    |                   |                    |                                       |                   |                         |                     | "address":"<IP address         |
|                    |                   |                    |                                       |                   |                         |                     | [192.168.1.10]>",              |
|                    |                   |                    |                                       |                   |                         |                     | "port":47808}                  |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Texas Instruments  | `CC2650STK`_      | `TI for sale`_     | Multi sensor: 3-axis accelerometer,   | device-bluetooth  |`CC2650STK repo`_        | Auto-provisionable  | TCP: {"name":"TI-CC2650<MAC    |
|                    |                   |                    | 3-axis gyroscope, 3-axis magnetometer,|                   |                         | by reference        | Address>","protocol":"MAC",    |
|                    |                   |                    | light, pressure, humidity,            |                   |                         | device service      | "path":"<MAC Address>"}        |
|                    |                   |                    | temperature                           |                   |                         | configuration       |                                |
|                    |                   |                    |                                       |                   |                         |                     |                                |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Bosch              | `XDK`_            | `Bosch for sale`_  | Multi sensor: 3-axis accelerometer,   | device-bluetooth  |`XDK repo`_              | Auto-provisionable  | TCP: {"name":"Bosch-XDK-<MAC   |
|                    |                   |                    | 3-axis gyroscope, 3-axis magnetometer,|                   |                         | by reference        | Address>","protocol":"MAC",    |
|                    |                   |                    | light, pressure, humidity,            |                   |                         | device service      | "path":"<MAC Address>"}        |
|                    |                   |                    | temperature                           |                   |                         | configuration       |                                |
|                    |                   |                    |                                       |                   |                         |                     |                                |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Texas Instruments  |`CC2540TDK-LIGHT`_ |`TI light for sale`_| BLE lighting control for LED          | device-bluetooth  |`CC2540TDK repo`_        | Auto-provisionable  | TCP:{" name":"TI-CC2540TDK-<MAC|
|                    |                   |                    | light products                        |                   |                         | by reference        | Address>","protocol":"MAC",    |
|                    |                   |                    |                                       |                   |                         | device service      | "path":"<MAC Address>"}        |
|                    |                   |                    |                                       |                   |                         | configuration       |                                |
|                    |                   |                    |                                       |                   |                         |                     |                                |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Fischertechnik     |`Punching Machine`_|`FT for sale`_      | XS motor (DC motor), Push button      | device-           |`FT repo`_               | Auto-provisionable  |Serial: {"name":"Fischertechnik |
|                    |                   |                    | (limit switch), Phototransistor,      | fischertechnic    |                         | by reference        |Punching Machine",              |
|                    |                   |                    | Lens tip lamp                         |                   |                         | device service      |"protocol":"OTHER",             | 
|                    |                   |                    |                                       |                   |                         | configuration       |"address":"","port":""}         |
|                    |                   |                    |                                       |                   |                         |                     |                                |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Dent Instruments   |`PS3037`_          |`Dent for sale`_    | Power meter: 3 line voltage,          | device-modbus     |`PS3037 repo`_           | Supports TCP        | TCP: {"name":"Power Scout",    |
|                    |                   |                    | current, and power draw               |                   |                         | Serial connection   | "method":"GET",                |
|                    |                   |                    |                                       |                   |                         | to device           | "protocol":"HTTP",             | 
|                    |                   |                    |                                       |                   |                         |                     | "address":"<IP address         |
|                    |                   |                    |                                       |                   |                         |                     | [192.168.1.10]>",              |
|                    |                   |                    |                                       |                   |                         |                     | "port":502}                    |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Automation Direct  |`DL05 PLC`_        |`DL05 for sale`_    | Modbus PLC, Eight built-in inputs     | device-modbus     |`DL05 repo`_             | Supports TCP        | TCP: {"name":"DL05",           |
|                    |                   |                    | and six built-in outputs              |                   |                         | connection          | "method":"GET",                |
|                    |                   |                    |                                       |                   |                         |                     | "protocol":"HTTP",             | 
|                    |                   |                    |                                       |                   |                         |                     | "address":"<IP address         |
|                    |                   |                    |                                       |                   |                         |                     | [192.168.1.10]>",              |
|                    |                   |                    |                                       |                   |                         |                     | "port":502}                    |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Automation Direct  |`GS1-10P5`_        |`AD for sale`_      | Power meter: 3 line voltage,          | device-modbus     |`GS1-10P5 repo`_         | Supports TCP        | TCP: {"name":"Variable speed   |
|                    |                   |                    | current, and power draw               |                   |                         | connection          | motor,                         |
|                    |                   |                    |                                       |                   |                         |                     | "protocol":"0THER",            |
|                    |                   |                    |                                       |                   |                         |                     | "address":"<IP address         |
|                    |                   |                    |                                       |                   |                         |                     | [192.168.1.10]>",              |
|                    |                   |                    |                                       |                   |                         |                     | "port":502}                    |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Datanab            |`MBUS_RTH_LCD`_    |`Datanab for sale`_ | Modbus Thermostat - Temperature       | device-modbus     |`MBUS_RTH_LCD repo`_     | Supports Serial     | Serial: {"name":"mbus-rth-lcd- |
|                    |                   |                    | F/C and Humidity                      |                   |                         | connection          | address",                      |
|                    |                   |                    |                                       |                   |                         |                     | "protocol":"0THER",            |
|                    |                   |                    |                                       |                   |                         |                     | "address":"<Serial/COM Port    |
|                    |                   |                    |                                       |                   |                         |                     | [/dev/ttyS5|COM6]>,            |
|                    |                   |                    |                                       |                   |                         |                     | <baud rate [9600]>,            |
|                    |                   |                    |                                       |                   |                         |                     | <data bits [8]>,               |
|                    |                   |                    |                                       |                   |                         |                     | <stop bits [1]>,               |
|                    |                   |                    |                                       |                   |                         |                     | <parity bits [1]>"}            |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Patlite            |`NHL-3FB1`_        |`Patlite for sale`_ | Three-tiered signal tower with a      | device-snmp       |`NHL-3FB1 repo`_         | Supports TCP        | TCP: {"name":"patlite-NHL-3FBL-|
|                    |                   |                    | buzzer nd E-mail sending capability   |                   |                         | connection          | address",                      |
|                    |                   |                    |                                       |                   |                         |                     | "protocol":"HTTP",             |
|                    |                   |                    |                                       |                   |                         |                     | "address":"<IP address         |
|                    |                   |                    |                                       |                   |                         |                     | [192.168.1.10]>",              |
|                    |                   |                    |                                       |                   |                         |                     | "port":161}                    |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+
| Moxa               |`iologik E2210`_   |`Moxa for sale`_    | SNMP Smart Ethernet Remote I/O        | device-snmp       |`iologik E2210 repo`_    | Supports TCP        | TCP: {"name":"moxa-e2210-      |
|                    |                   |                    | with 12 DIs, 8 DOs                    |                   |                         | connection          | address",                      |
|                    |                   |                    |                                       |                   |                         |                     | "protocol":"HTTP",             |
|                    |                   |                    |                                       |                   |                         |                     | "address":"<IP address         |
|                    |                   |                    |                                       |                   |                         |                     | [192.168.1.10]>",              |
|                    |                   |                    |                                       |                   |                         |                     | "port":161}                    |
+--------------------+-------------------+--------------------+---------------------------------------+-------------------+-------------------------+---------------------+--------------------------------+

















