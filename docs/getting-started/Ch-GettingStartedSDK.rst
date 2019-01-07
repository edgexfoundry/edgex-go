##################################
Writing a Device Service in Go
##################################

The EdgeX Device Service SDK helps developers quickly create new device connectors for EdgeX because it provides the common scaffolding that each Device Service needs to have.  The scaffolding provides a pattern for provisioning devices.  It provides common template code to receive and react to command (a.k.a. actuation) requests.  Finally, the scaffolding provides the common code to help get the data coming from the sensor into EdgeX Core Data (often referred to as data ingestion).  With the SDK, developers are left to focus on the code that is specific to the communications with the device via the protocol of the device.

In this guide, you create a simple device service that generates a random number in place of getting data from an actual sensor.  In this way, you get to explore some of the scaffolding and work necessary to complete a device service without actually having a device to talk to.

====================
Install dependencies
====================

Creating a device service requires some small amounts of programming in Go.  Go Lang (version 1.10 or better) must be installed on your system to complete this lab.  Follow the instructions below to install Go if it is not already installed on your platform.
    https://golang.org/doc/install


Glide is a tool for managing 3rd party packages for Go.   EdgeX uses glide to retrieve and manage the 3rd party libraries used in the system.  Follow the instructions below to get Glide for your environment.
    https://github.com/Masterminds/glide


You will also need a Git tool in order to pull the Device Service Go SDK code from the EdgeX Foundry Git repository.  Follow the instructions below to install Git for your platform.
    https://git-scm.com/book/en/v2/Getting-Started-Installing-Git


You will also need a “make” program.  On Ubuntu Linux environments, this can be accomplished with the following command::

    sudo apt install build-essential

Finally, you will need a simple text editor (or Go Lang IDE).

===============================
Get the EdgeX Device SDK for Go
===============================

In this step, you create a folder on your file system where you will download the :doc:`./Ch-DeviceSDK`, then you pull the SDK to your system, and finally create the new EdgeX Device Service from the SDK templating code.

#. Create a collection of nested folders, ~/go/src/github.com/edgexfoundry on your file system.  This folder will eventually hold your new Device Service.  In linux, this can be done with a single mkdir (with -p switch) command::

    mkdir -p ~/go/src/github.com/edgexfoundry

#. In a terminal window, change directories to the folder you created::

    cd ~/go/src/github.com/edgexfoundry

#. Type the following command to pull down the EdgeX Device Service SDK in Go (there is also a Device Service SDK in C)::

    git clone https://github.com/edgexfoundry/device-sdk-go.git

   .. image:: EdgeX_GettingStartedSDKClone3.png
    
#. Rename the cloned **device-sdk-go** to **device-simple**. In this step, you are renaming the folder to the name you want to give your new Device Service.  Standard practice in EdgeX is to prefix the name of a Device Service with **device-** ::

    mv device-sdk-go/ device-simple

   .. image:: EdgeX_GettingStartedSDKClone4.png
    

=====================================
Starting a new Device Service project
=====================================

The device-sdk-go comes with example code to create a new device service.  In this step, you move a copy of the example code to use in your new service.

#. Locate the cmd and driver folders in the /example directory of the device-simple folder.  Move both the cmd and driver folders to the root of the device-simple folder (you can use command line or file system explorer tool to move this folder).

   .. image:: EdgeX_GettingStartedSDKProject1.png

#. Search for all files with references to “device-sdk-go” in the device-simple folder (and its subfolders).  If you are using Linux or Unix, you can use the following command to find all files with “device-sdk-go” in the device-simple folder::

    grep -irl “device-sdk-go”

   You should find the following files have “device-sdk-go” references

   .. image:: EdgeX_GettingStartedSDKProject3.png

#. Edit each of the files above and change “device-sdk-go” to “device-simple”.  You can ignore any of the “.git” files.  If you are on a linux system, you can use this one command to replace all the occurances::

    find .  -type f | xargs sed -i  's/device-sdk-go/device-simple/g'

#. Edit the main.go file in the cmd/device-simple folder.  Modify the import statements to remove “/example” from the paths in the import statements.  Save the file when done editing.

   .. image:: EdgeX_GettingStartedSDKProject4.png

#. Edit the Makefile in the root of the /device-simple folder.  In the Makefile, change all of the example/cmd/device-simple to be just cmd/device-simple (there are 3 such references in this directory.

   .. image:: EdgeX_GettingStartedSDKProject5.png

#. Remove the now empty /device-simple/example directory::

    rmdir example/

=========================
Build your Device Service
=========================

In order to ensure that the code you have moved and updated still works, build the current Device Service.

#. In a terminal window, change directories to the device-simple folder (the folder contains the Makefile).
#. Get the associated 3rd party libraries used in the service by issuing the command::

    make prepare

   It will take a few minutes for Go Glide to download all the 3rd party dependencies.

#. Now build the service by issuing the command::

    make build

   .. image:: EdgeX_GettingStartedSDKBuild1.png

#. If there are no errors, your service is ready for you to add customizations to generate data values as if there was a sensor attached.  If there are errors, retrace your steps to correct the error and try to build again.  Ask you instructor for help in finding the issue if you are unable to locate it given the error messages you receive from the build process.

   .. image:: EdgeX_GettingStartedSDKBuild2.png

=============================
Customize your Device Service
=============================

The Device Service you are creating isn’t going to talk to a real device.  Instead, it is going to simply generate a random number in place of where the service would make a call to get sensor data from the actual device.  By so doing, you see where the EdgeX Device Service would make a call to a local device (via its protocol and device drivers under the covers) in order to provide EdgeX with its sensor readings.

#. Locate the simpledriver.go file in the /driver folder and open it with your favorite editor.

   .. image:: EdgeX_GettingStartedSDKCode1.png

#. In the import( ) area at the top of the file, add “math/rand” under “time”.

   .. image:: EdgeX_GettingStartedSDKCode2.png

#. Next, locate the HandleReadCommands() function in this file.  Notice the following line of code in this file::

    cv, _ := ds_models.NewBoolValue(&reqs[0].RO, now, s.switchButton)

   .. image:: EdgeX_GettingStartedSDKCode3.png

#. Replace that line of code with the following line of code::

    cv, _ := ds_models.NewInt32Value(&reqs[0].RO, now, int32(rand.Intn(100)))

   .. image:: EdgeX_GettingStartedSDKCode4.png

#. What this line of code does is generate an integer (between 0 and 100) and uses that as the value the Device Service sends to EdgeX – mimicking the collection of data from a real device.  It is here that the Device Service would normally capture some sensor reading from a device and send the data to EdgeX.  The line of code you just added is where you’d need to do some customization work to talk to the sensor, get the sensor's latest sensor values and send them into EdgeX.

#. Save the simpledriver.go file

============================
Creating your Device Profile
============================

A Device Profile is a YAML file that describes a class of device to EdgeX.  General characteristics about the type of device, the data these devices provide, and how to command the device is all provided in a Device Profile.  Device Services use the Device Profile to understand what data is being collected from the Device (in some cases providing information used by the Device Service to know how to communicate with the device and get the desired sensor readings).  A Device Profile is needed to describe the data that will be collected from the simple random number generating Device Service.

#. Explore the files in the cmd/device-simple/res folder.  Take note of the example Device Profile YAML file that is already there (Simple-Driver.yml).  You can explore the contents of this file to see how devices are represented by YAML.  In particular, note how fields or properties of a sensor are represented by “deviceResources”.  Command to be issued to the device are represented by “commands”.

#. Download this :download:`random-generator-device.yaml <random-generator-device.yaml>` into the cmd/device-simple/res folder.  

You can open random-generator-device.yaml in a text editor.  In this Device Profile, you are suggesting that the device you are describing to EdgeX has a single property (or deviceResource) which EdgeX should know about - in this case, the property is the “randomnumber”.  Note how the deviceResource is typed.

    In more real world IoT situations, this deviceResource list could be extensive and could be filled with all different types of data.

    Note also how the Device Profile describes REST commands that can be used by others to call on (or “get”) the random number from the Device Service.   

===============================
Configuring your Device Service
===============================

You will now update the configuration for your new Device Service – changing the port it operates on (so as not to conflict with other Device Services), altering the scheduled times of when the data is collected from the Device Service (every 10 seconds), and setting up the initial provisioning of the random number generating device when the service starts.

* Downlod this :download:`configuration.toml <configuration.toml>` to the cmd/device-simple/res folder (this will overwrite an existing file – that’s ok).  

===========================
Rebuild your Device Service
===========================

Just as you did before, you are ready to build the device-simple service – creating the executable program that is your Device Service.

#. As you did previously, in a terminal window, change directories to the base device-simple folder (containing the Makefile).

#. Build the Device Service by issuing the command::

    make build

   .. image:: EdgeX_GettingStartedSDKRebuild1.png

#. If there are no errors, your service has now been created and is available in the cmd/device-simple folder (look for the device-simple file).

=======================
Run your Device Service
=======================

Allow your newly created Device Service, which was formed out of the Device Service Go SDK, to create sensor mimicking data which it then sends to EdgeX.

#. Per the :doc:`./Ch-GettingStartedUsers` guide, use Docker Compose to start all of EdgeX.  From the folder containing the docker-compose file, start EdgeX with a call to::

    docker-compose up -d

#. In a terminal window, change directories to the device-simple’s cmd/device-simple folder.  You should find the executable device-simple there.

   .. image:: EdgeX_GettingStartedSDKRun1.png

#. Execute the Device Service with the command ./device-simple (as shown below)

   .. image:: EdgeX_GettingStartedSDKRun2.png

   This will start the service and it will immediate start to display log entries in the terminal.

#. Using a browser, use the following URL to see the Event/Reading data that the service is generating and sending into EdgeX.

   http://localhost:48080/api/v1/event/device/RandNum-Device-01/100

   .. image:: EdgeX_GettingStartedSDKRun3.png

   This request asks for the last 100 Events/Readings from Core Data associated to the RandNum-Device-01.

   **Note**: If you are running the other EdgeX services somewhere other than localhost, use that hostname in the above URL.