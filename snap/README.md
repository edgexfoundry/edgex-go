# EdgeX Foundry Core Snap
This folder contains snap packaging for the EdgeX Foundry reference implementation.

The snap contains Consul, MongoDB, all of the EdgeX Go-based micro services from
this repository, and three Java-based services, support-rulesengine and support-
scheduler, and device-virtual, as well as Vault, Kong, Cassandra, and the two go-based 
security micro services. The snap also contains a single OpenJDK JRE used to run 
the Java-based services.

The project maintains a rolling release of the snap on the `edge` channel that is rebuilt and published at least once daily through the jenkins jobs setup for the EdgeX project. You can see the jobs run [here](https://jenkins.edgexfoundry.org/view/Snap/) specifically looking at the `edgex-go-snap-{branch}-stage-snap`.

The snap currently supports running on both `amd64` and `arm64` platforms. Once the project no longer depends on MongoDB as it's primary database, the snap should start working on `armhf` platforms that support snaps as well.

## Installation

### Installing snapd
The snap can be installed on any system that supports snaps. You can see how to install 
snaps on your system [here](https://docs.snapcraft.io/t/installing-snapd/6735).

However for full security confinement, the snap should be installed on an 
Ubuntu 16.04 LTS or later Desktop or Server, or a system running Ubuntu Core 16 or later.

### Installing EdgeX Foundry as a snap
The snap is published in the snap store at https://snapcraft.io/edgexfoundry.
You can see the current revisions available for your machine's architecture by running the command:

```
$ snap info edgexfoundry
```

The snap can be installed using `snap install`. To install the snap from the edge channel:

```
$ sudo snap install edgexfoundry --edge
```

You can specify install specific releases using the `--channel` option. For example to install the california release of the snap:

```
$ sudo snap install edgexfoundry --channel=california
```

Lastly, on a system supporting it, the snap may be installed using GNOME (or Ubuntu) Software Center by searching for `edgexfoundry`.

**Note** - the snap has only been tested on Ubuntu 16.04 LTS Desktop/Server and Ubuntu Core 16.

### Configuration
The `hardware-observe`, `process-control`, `mount-observe`, and `system-observe` snap interfaces needs to be
connected after installation using the following commands:

```
$ snap connect edgexfoundry:hardware-observe
$ snap connect edgexfoundry:process-control
$ snap connect edgexfoundry:system-observe
$ snap connect edgexfoundry:mount-observe
```


**Note** - these interface will be connected automatically after https://forum.snapcraft.io/t/edgexfoundry-auto-connections-assertions-request/7920/1 has been processed.

**Note** - the process-control interface will be dropped after all of the services are supported as proper daemons in the snap.

## Starting/Stopping EdgeX
To start all the EdgeX microservices, run the following command as root:

`$ edgexfoundry.start-edgex`

To stop all the EdgeX microservices, run the following command as root:

`$ edgexfoundry.stop-edgex`

**WARNING** - don't start the EdgeX snap on a system which is already running mongoDB or Consul.

### Enabling/Disabling service startup
It's possible to change which services are started by the `start-edgex` script by
editing a file called `edgex-services-env` which can be found in the directory `/var/snap/edgexfoundry/current` (aka $SNAP_DATA).

**Note** - after all services have been converted to daemons in the snap this file will become obselete.

## Limitations

  * None of the services are actually defined as such in snapcraft.yaml, instead shell-scripts are used to start and stop the EdgeX microservices and dependent services such as consul and mongo. This is being tracked at https://github.com/edgexfoundry/edgex-go/issues/485

## Building

The source for the snap is inside this repo (the `edgex-go` repo), so the first step for all build methods involves cloning this repository:

```bash
$ git clone https://github.com/edgexfoundry/edgex-go
$ cd edgex-go
```

The `snapcraft` tool is used to actually build the snap. There are a few different ways to use it, depending on what OS you are building on.

### Building on Ubuntu 16.04

The easiest way to build the snap is on an Ubuntu 16.04 Classic installation where you can use `snapcraft` directly:

```bash
$ snapcraft
```

This will produce a binary snap package called `edgexfoundry_<latest version>_<arch>.snap`, which can be installed locally with the `--dangerous` flag:

```
sudo snap install --dangerous edgexfoundry*.snap
```

### Building with LXD containers

Alternatively, you can use snapcraft with it's own container/build VM using `cleanbuild`. This requires installing LXD as documented (here)[https://docs.snapcraft.io/t/clean-build-using-lxc].

```bash
$ snapcraft cleanbuild
```

### Building using docker

Lastly, you can use docker with the `snapcraft` docker image to build the snap. You will build the snap inside a container, using the repository on your host filesystem as a volume mapped into the container. This lets us access the build artifacts and makes it easier to iterate on builds by sharing the intermediate contents of the build between runs of `snapcraft`.

```bash
$ docker run -it -v"$PWD":/build snapcore/snapcraft:stable bash -c "apt update && cd /build && snapcraft"
```

If the build fails and you need to make changes and re-build, you can use snapcraft commands such as `snapcraft clean`, etc. inside the container like so:

```bash
$ docker run -it -v"$PWD":/build snapcore/snapcraft:stable bash -c "apt update && cd /build && snapcraft clean && snapcraft"
```
