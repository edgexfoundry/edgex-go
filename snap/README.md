# EdgeX Foundry Core Snap
[![snap store badge](https://raw.githubusercontent.com/snapcore/snap-store-badges/master/EN/%5BEN%5D-snap-store-black-uneditable.png)](https://snapcraft.io/edgexfoundry)

This folder contains snap packaging for the EdgeX Foundry reference implementation.

The snap contains Consul, MongoDB, Redis, and all of the EdgeX Go-based micro services from
this repository, device-virtual, as well as Vault, Kong, PostgreSQL. The snap also contains a
single OpenJDK JRE used to run the legacy support-rulesengine (deprecated).

The project maintains a rolling release of the snap on the `edge` channel that is rebuilt and published at least once daily through the jenkins jobs setup for the EdgeX project.

The snap currently supports running on both `amd64` and `arm64` platforms.

## Installation

### Installing snapd
The snap can be installed on any system that supports snaps. You can see how to install 
snaps on your system [here](https://snapcraft.io/docs/installing-snapd).

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

You can specify install specific releases using the `--channel` option. For example to install the Fuji release of the snap:

```bash
$ sudo snap install edgexfoundry --channel=fuji

```

Lastly, on a system supporting it, the snap may be installed using GNOME (or Ubuntu) Software Center by searching for `edgexfoundry`.

**Note** - the snap has only been tested on Ubuntu Desktop/Server versions 18.04 and 16.04, as well as Ubuntu Core versions 16 and 18.

**WARNING** - don't install the EdgeX snap on a system which is already running one of the included services (e.g. Consul, Redis, Vault, ...).

## Using the EdgeX snap

Upon installation, the following EdgeX services are automatically and immediately started:

* consul
* redis
* core-data
* core-command
* core-metadata
* security-services (see [note below](https://github.com/edgexfoundry/edgex-go/tree/master/snap#security-services))

The following services are disabled by default:

* app-service-configurable (required for Kuiper and support-rulesengine)
* device-virtual
* kuiper
* support-logging
* support-notifications
* support-rulesengine (deprecated)
* support-scheduler
* sys-mgmt-agent

Any disabled services can be enabled and started up using `snap set`:

```bash
$ sudo snap set edgexfoundry support-notifications=on
```

To turn a service off (thereby disabling and immediately stopping it) set the service to off:

```bash
$ sudo snap set edgexfoundry support-notifications=off
```

All services which are installed on the system as systemd units, which if enabled will automatically start running when the system boots or reboots.

### Configuring individual services

All default configuration files are shipped with the snap inside `$SNAP/config`, however because `$SNAP` isn't writable, all of the config files are copied during snap installation (specifically during the install hook, see `snap/hooks/install` in this repository) to `$SNAP_DATA/config`.

Note - as the core-config-seed was removed as part of the Geneva release, services self-seed their configuration on startup. This means that if a service is
started by default in the snap, the only way to change configuration is to use the Consul UI or [kv REST API](https://www.consul.io/api/kv.html). Services that
aren't started by default (see above) *will* pickup any changes made to their config files when enabled.

```bash
$ sudo snap restart edgexfoundry
```

### Viewing logs

Currently, all log files for the snap's can be found inside `$SNAP_COMMON`, which is usually `/var/snap/edgexfoundry/common`. Once all the services are supported as daemons, you can also use `sudo snap logs edgexfoundry` to view logs.

Additionally, logs can be viewed using the system journal or `snap logs`. To view the logs for all services in the edgexfoundry snap use:

```bash
$ sudo snap logs edgexfoundry
```

Individual service logs may be viewed by specifying the service name:

```bash
$ sudo snap logs edgexfoundry.consul
```

Or by using the systemd unit name and `journalctl`:

```bash
$ journalctl -u snap.edgexfoundry.consul
```

### Security services

Currently, the security services are enabled by default. The security services consitute the following components:

 * Kong
 * PostgreSQL
 * Vault
 * security-secrets-setup
 * security-secretstore-setup
 * security-proxy-setup

Vault is used for secret management, and Kong is used as an HTTPS proxy for all the services.

Kong can be disabled by using the following command:

```bash
$ sudo snap set edgexfoundry security-proxy=off
```

Vault can be also be disabled, but doing so will also disable Kong, as it depends on Vault. Thus the following command will disable both:

```bash
$ sudo snap set edgexfoundry security-secret-store=off
```
**Note** - Kong is currently not supported in the snap when installed on an arm64-based device, so it will be disabled on install.

## Limitations

[See the GitHub issues with label snap for current issues.](https://github.com/edgexfoundry/edgex-go/issues?q=is%3Aopen+is%3Aissue+label%3Asnap)

## Building

The snap is built with [snapcraft](https://snapcraft.io), and the snapcraft.yaml recipe is located within `edgex-go`, so the first step for all build methods involves cloning this repository:

```bash
$ git clone https://github.com/edgexfoundry/edgex-go
$ cd edgex-go
```

### Installing snapcraft

There are a few different ways to install snapcraft and use it, depending on what OS you are building on. However after building, the snap can only be run on a Linux machine (either a VM or natively). To install snapcraft on a Linux distro, first [install support for snaps](https://snapcraft.io/docs/installing-snapd), then install snapcraft as a snap with:

```bash
$ sudo snap install snapcraft
```

(note you will be promted to acknowledge you are installing a classic snap - use the `--classic` flag to acknowledge this)

**Note** - currently the snap doesn't support cross-compilation, and must be built natively on the target architecture. Specifically, to support cross-compilation the kong/lua parts must be modified to support cross-compilation. The openresty part uses non-standard flags for handling cross-compiling so all the flags would have to manually passed to build that part. Also luarocks doesn't seem to easily support cross-compilation, so that would need to be figured out as well.

#### Running snapcraft on MacOS

To install snapcraft on MacOS, see [this link](https://snapcraft.io/docs/install-snapcraft-on-macos). After doing so, follow in the below build instructions for "Building with multipass"

#### Running snapcraft on Windows

To install snapcraft on Windows, you will need to run a Linux VM and follow the above instructions to install snapcraft as a snap. Note that if you are using WSL, only WSL2 with full Linux kernel support will work - you cannot use WSL with snapcraft and snaps. If you like, you can install multipass to launch a Linux VM if your Windows machine has Windows 10 Pro or Enterprise with Hyper-V support. See this [forum post](https://discourse.ubuntu.com/t/installing-multipass-for-windows/9547) for more details.

### Building with multipass

The easiest way to build the snap is using the multipass VM tool that snapcraft knows to use directly. After [installing multipass](https://multipass.run), just run 

```bash
$ snapcraft
```

### Building with LXD containers

Alternatively, you can instruct snapcraft to use LXD containers instead of multipass VM's. This requires installing LXD as documented [here](https://snapcraft.io/docs/build-on-lxd).

```bash
$ snapcraft --use-lxd
```

Note that if you are building on non-amd64 hardware, snapcraft won't be able to use it's default LXD container image, so you can follow the next section to create an LXD container to run snapcraft in destructive-mode natively in the container.

### Building inside external container/VM using native snapcraft

Finally, snapcraft can be run inside a VM, container or other similar build environment to build the snap without having snapcraft manage the environment (such as in a docker container where snaps are not available, or inside a VM launched from a build-farm without using nested VM's). 

This requires creating an Ubuntu 18.04 environment and running snapcraft (from the snap) inside the environment with `--destructive-mode`. 

#### LXD

Snaps run inside LXD containers just like they do outside the container, so all you need to do is launch an Ubuntu 18.04 container, install snapcraft and run snapcraft like follows:

```bash
$ lxc launch ubuntu:18.04 edgex
Creating edgex
Starting edgex
$ lxc exec edgex /bin/bash
root@edgex:~# sudo apt update && sudo apt install snapd squashfuse git -y
root@edgex:~# sudo snap install snapcraft --classic
root@edgex:~# git clone https://github.com/edgexfoundry/edgex-go
root@edgex:~# cd edgex-go && snapcraft --destructive-mode
```

#### Docker

Snapcraft is smart enough to detect when it is running inside a docker container specifically, to the point where no additional arguments are need to snapcraft when it is run inside the container. For example, the upstream snapcraft docker image can be used (only on x86_64 architectures unfortunately) like so:

```bash
$ docker run -it -v"$PWD":/build snapcore/snapcraft:stable bash -c "apt update && cd /build && snapcraft"
```

Note that if you are building your own docker image, you can't run snapd inside the container, and so to install snapcraft, the docker image must download the snapcraft snap and extract it as if it was installed normally inside `/snap` (same goes for the `core` and `core18` snaps). This is done by the Linux Foundation Jenkins server for the project's CI and you can see an example of that [here](https://github.com/edgexfoundry/ci-management/blob/master/shell/edgexfoundry-snapcraft.sh). The upstream docker image also does this, but only for x86_64 architectures.

#### Multipass / generic VM

To use multipass to create an Ubuntu 18.04 environment suitable for building the snap (i.e. when running natively on windows):

```bash
$ multipass launch bionic -n edgex-snap-build
$ multipass shell edgex-snap-build
multipass@ubuntu:~$ git clone https://github.com/edgexfoundry/edgex-go
multipass@ubuntu:~$ cd edgex-go
multipass~ubuntu:~$ sudo snap install snapcraft --classic
multipass~ubuntu:~$ snapcraft --destructive-mode
```

The process should be similar for other VM's such as kvm, VirtualBox, etc. where you create the VM, clone the git repository, then install snapcraft as a snap and run with `--destructive-mode`. 

### Developing the snap

After building the snap from one of the above methods, you will have a binary snap package called `edgexfoundry_<latest version>_<arch>.snap`, which can be installed locally with the `--devmode` flag:

```bash
$ sudo snap install --devmode edgexfoundry*.snap
```

In addition, if you are using snapcraft with multipass VM's, you can speedup development by not creating a *.snap file and instead running in "try" mode . This is done by running `snapcraft try` which results in a `prime` folder placed in the root project directory that can then be "installed" using `snap try`. For example:

```bash
$ snapcraft try # produces prime dir instead of *.snap file
...
You can now run `snap try /home/ubuntu/go/src/github.com/edgexfoundry/edgex-go/prime`.
$ sudo snap try --devmode prime # snap try works the same as snap install, but expects a directory
edgexfoundry 1.0.0-20190513+0620a8d1 mounted from /home/ubuntu/go/src/github.com/edgexfoundry/edgex-go/prime
$
```

#### Interfaces

After installing the snap, you will need to connect interfaces and restart the snap. The snap needs the `hardware-observe`, `mount-observe`, and `system-observe` interfaces connected. These are automatically connected using snap store assertions when installing from the store, but when developing the snap and installing a revision locally, use the following commands to connect the interfaces:

```bash
$ sudo snap connect edgexfoundry:hardware-observe
$ sudo snap connect edgexfoundry:system-observe
$ sudo snap connect edgexfoundry:mount-observe
```

After connecting these restart the services in the snap with:

```bash
$ sudo snap restart edgexfoundry
```

