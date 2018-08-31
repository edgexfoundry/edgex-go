# EdgeX Foundry Core Snap
This project contains snap packaging for the EgdeX Foundry reference implementation.

The snap contains consul, mongodb, all of the EdgeX Go-based micro services from
this repository, and three Java-based services, support-notifications and support-
scheduler, and device-virtual. The snap also contains a single OpenJDK JRE used to
run the Java-based services.

## Installation Requirements
The snap can be installed on any system running snapd, however for full confinement,
the snap must be installed on an Ubuntu 16.04 LTS or later Desktop or Server, or a
system running Ubuntu Core 16.

## Installation
There are amd64 and arm64 versions of both releases available in the store.  You can
see the revisions available for your machine's architecture by running the command:

`$ snap info edgexfoundry`

The snap can be installed using this command:

`$ sudo snap install edgexfoundry --channel=california/edge`

**Note** - this snap has only been tested on Ubuntu 16.04 LTS Desktop/Server and Ubuntu Core 16.

## Configuration
The hardware-observe, process-control, mount-observe, and system-observe snap interfaces needs to be
connected after installation using the following commands:

`$ snap connect edgexfoundry-core:hardware-observe core:hardware-observe`

`$ snap connect edgexfoundry-core:process-control core:process-control`

`$ snap connect edgexfoundry-core:system-observe core:system-observe`

`$ snap connect edgexfoundry-core:system-observe core:mount-observe`


## Starting/Stopping EdgeX
To start all the EdgeX microservices, use the following command:

`$ edgexfoundry.start-edgex`

To stop all the EdgeX microservices, use the following command:

`$ edgexfoundry.stop-edgex`

**WARNING** - don't start the EdgeX snap on a system which is already running mongoDB or Consul.

### Enabling/Disabling service startup
It's possible to a effect which services are started by the start-edgex script by
editing a file called `edgex-services-env` which can be found in the directory `/var/snap/edgexfoundry/current` (aka $SNAP_DATA).

**Note** - this file is created by the start-edgex script, so the script needs to be run at least once to copy the default version into place.

## Limitations

  * none of the services are actually defined as such in snapcraft.yaml, instead shell-scripts are used to start and stop the EdgeX microservices and dependent services such as consul and mongo.

  * some of the new Go-based core services (export-*) currently don't load configuration from Consul

  * the new Go-based export services don't generate local log files

## Building

This snap can be built on an Ubuntu 16.04 LTS system:

 * install snapcraft
 * clone this git repo
 * cd edgex-core-snap
 * snapcraft

This should produce a binary snap package called edgex-core-snap_<latest version>_<arch>.snap.
