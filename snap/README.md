# EdgeX Foundry Core Snap
[![snap store badge](https://raw.githubusercontent.com/snapcore/snap-store-badges/master/EN/%5BEN%5D-snap-store-black-uneditable.png)](https://snapcraft.io/edgexfoundry)

This folder contains snap packaging for the EdgeX Foundry reference implementation.

The snap contains Consul, MongoDB, all of the EdgeX Go-based micro services from
this repository, and two Java-based services support-scheduler, and device-virtual,
as well as Vault, Kong, Cassandra, and the two go-based security micro services.
The snap also contains a single OpenJDK JRE used to run the Java-based services.

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

```bash
$ snap info edgexfoundry
```

The snap can be installed using `snap install`. To install the snap from the edge channel:

```bash
$ sudo snap install edgexfoundry --edge
```

You can specify install specific releases using the `--channel` option. For example to install the Delhi release of the snap:

```bash
$ sudo snap install edgexfoundry --channel=delhi
```

Lastly, on a system supporting it, the snap may be installed using GNOME (or Ubuntu) Software Center by searching for `edgexfoundry`.

**Note** - the snap has only been tested on Ubuntu 16.04 LTS Desktop/Server and Ubuntu Core 16.

### Configuration
The `hardware-observe`, `process-control`, `mount-observe`, and `system-observe` snap interfaces needs to be
connected after installation using the following commands:

```bash
$ sudo snap connect edgexfoundry:hardware-observe
$ sudo snap connect edgexfoundry:process-control
$ sudo snap connect edgexfoundry:system-observe
$ sudo snap connect edgexfoundry:mount-observe
```

After connecting these restart the services in the snap with:

```bash
$ sudo snap restart edgexfoundry
```

**Note** - these interface will be connected automatically after https://forum.snapcraft.io/t/edgexfoundry-auto-connections-assertions-request/7920/1 has been processed. After this is done, then connecting these interfaces will only be necessary during development. See the section below on building the snap for more details.

**Note** - the process-control interface will be dropped after all of the services are supported as proper daemons in the snap.

## Using the EdgeX snap
To start all the EdgeX microservices, run the following command using sudo:

`$ edgexfoundry.start-edgex`

To stop all the EdgeX microservices, run the following command using sudo:

`$ edgexfoundry.stop-edgex`

**WARNING** - don't start the EdgeX snap on a system which is already running mongoDB or Consul.

### Enabling/Disabling service startup
It's possible to change which services are started by the `start-edgex` script by
editing a file called `edgex-services-env` which can be found in the directory `/var/snap/edgexfoundry/current` (aka $SNAP_DATA).

**Note** - after all services have been converted to daemons in the snap this file will become obselete and all service management will be done with `snap set`. This document will be updated to reflect this...

### Configuring individual services

All default configuration files are shipped with the snap inside `$SNAP/config`, however because `$SNAP` isn't writable, all of the config files are copied during snap installation (specifically during the install hook, see `snap/hooks/install` in this repository) to `$SNAP_DATA/config`. The configuration files in `$SNAP_DATA` may then be modified.

### Viewing logs

Currently, all log files for the snap's can be found inside `$SNAP_COMMON`, which is usually `/var/snap/edgexfoundry/common`. Once all the services are supported as daemons, you can also use `sudo snap logs edgexfoundry` to view logs.

### Enabling security services

Currently, the security services are enabled by default. The security services consitute the following components:

 * Vault
 * Cassandra
 * Kong
 * vault-worker (from [security-secret-store](https://github.com/edgexfoundry/security-secret-store))
 * kong-worker (from [security-api-gateway](https://github.com/edgexfoundry/security-api-gateway/))

All of the services are controlled by the same variable `SECURITY` in the `edgex-services-env` file. You can either turn all of them on, or turn all of them off. 

When security is enabled (which is what happens by default), Consul is secured using Vault for secret management, and Kong is used as an HTTPS proxy for all the services. The HTTPS keys for Kong are generated at runtime, and placed in `$SNAP_DATA/kong/keys` and the keys for Vault are placed in `$SNAP_DATA/vault/pki`. Kong needs a database to manage itself, and can use either Postgres or Cassandra. Because Postgres cannot run inside of a snap due to issues running as root (currently all snap services must run as root, see [this post](https://forum.snapcraft.io/t/multiple-users-and-groups-in-snaps/1461) for details), we use Cassandra. Cassandra is currently a service, so to fully disable the security services, you must also disable Cassandra using:

```
$ sudo snap stop --disable edgexfoundry.cassandra
```

## Limitations

  * None of the services are actually defined as such in snapcraft.yaml, instead shell-scripts are used to start and stop the EdgeX microservices and dependent services such as consul and mongo. This is being tracked at https://github.com/edgexfoundry/edgex-go/issues/485

## Building

The source for the snap is inside this repo (the `edgex-go` repo), so the first step for all build methods involves cloning this repository:

```bash
$ git clone https://github.com/edgexfoundry/edgex-go
$ cd edgex-go
```

The `snapcraft` tool is used to actually build the snap. There are a few different ways to use it, depending on what OS you are building on.

**Note** - currently the snap doesn't support cross-compilation, and must be built natively on the target architecture. Specifically, to support cross-compilation the kong/lua parts must be modified to support cross-compilation. The openresty part uses non-standard flags for handling cross-compiling so all the flags would have to manually passed to build that part. Also luarocks doesn't seem to easily support cross-compilation, so that would need to be figured out as well.

### Building on Ubuntu 16.04

The easiest way to build the snap is on an Ubuntu 16.04 Classic installation where you can use `snapcraft` directly:

```bash
$ snapcraft
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

### Developing the snap

After building the snap from one of the above methods, you will have a binary snap package called `edgexfoundry_<latest version>_<arch>.snap`, which can be installed locally with the `--devmode` flag:

```bash
sudo snap install --devmode edgexfoundry*.snap
```

After installing the snap, you will need to connect interfaces and restart the snap as explained in the previous section on [Configuration](https://github.com/edgexfoundry/edgex-go/blob/master/snap/README.md#configuration).
