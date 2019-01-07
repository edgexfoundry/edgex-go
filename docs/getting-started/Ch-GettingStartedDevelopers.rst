##############################
Getting Started - Developers
##############################

============
Introduction
============

These instructions are for Developers to obtain and run EdgeX Foundry.  (Users should read: :doc:`../Ch-GettingStartedUsers`) 

EdgeX Foundry is a collection of more than a dozen microservices that can be deployed to provide a minimal edge platform capability.  EdgeX Foundry consists of a collection of microservices and SDK tools.  The microservices and SDKs are mostly written in Go or C with some legacy servcies written in Java (EdgeX was originally written in Java).  These documentation pages provide a developer with the information and instructions to get and run EdgeX Foundry in either the Go or Java (for legacy services) development environments.

=============
What You Need
=============

**Hardware**

EdgeX Foundry is an operating system (OS)-agnostic and hardware (HW)-agnostic edge software platform. Minimum platform requirements are being established. At this time use the following recommended characteristics:

* Memory:  minimum of 1 GB 
* Hard drive space:  minimum of 3 GB of space to run the EdgeX Foundry containers, but you may want more depending on how long sensor and device data is retained
* OS: EdgeX Foundry has been run successfully on many systems including, but not limited to the following systems
        * Windows (ver 7 - 10)
        * Ubuntu Desktop (ver 14-16)
        * Ubuntu Server (ver 14)
        * Ubuntu Core (ver 16)
        * Mac OS X 10

**Software**

Developers will need to install the following software in order to get, run and develop EdgeX Foundry microservices:

**git** - a free and open source version control (SVC) system used to download (and upload) the EdgeX Foundry source code from the project's GitHub repository.  See https://git-scm.com/downloads for download and install instructions.  Alternative tools (Easy Git for example) could be used, but this document assumes use of git and leaves how to use alternative SVC tools to the reader.

**MongoDB** - by default, EdgeX Foundry uses MongoDB as the persistence mechanism for sensor data as well as metadata about the devices/sensors that are connected.  See https://www.mongodb.com/download-center?jmp=nav#community for download and installation instructions.  As an alternative to installing MongoDB directly, you can use a MongoDB on another server or in the cloud.  This document will explain how to setup MongoDB for use with your development environment.

.. toctree::
   :maxdepth: 1

   Ch-GettingStartedGoDevelopers
   Ch-GettingStartedJavaDevelopers

  
