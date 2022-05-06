# EdgeX Platform Snap
[![edgexfoundry](https://snapcraft.io/edgexfoundry/badge.svg)][edgexfoundry]

This directory contains the snap packaging of the EdgeX platform snap containing all reference core services along with several other security, supporting, application, and device services.

The platform snap is built automatically and published on the Snap Store as [edgexfoundry].

For usage instructions, please refer to Platform Snap section in [Getting Started using Snaps][docs].

## Limitations

[See the GitHub issues with label snap for current issues.](https://github.com/edgexfoundry/edgex-go/issues?q=is%3Aopen+is%3Aissue+label%3Asnap)

## Building

The snap is built with [snapcraft](https://snapcraft.io), and the snapcraft.yaml recipe is located within `edgex-go`, 
so the first step for all build methods involves cloning this repository:

```bash
git clone https://github.com/edgexfoundry/edgex-go
cd edgex-go
```

### Installing snapcraft

There are a few different ways to install snapcraft and use it, depending on what OS you are building on. However after building, 
the snap can only be run on a Linux machine (either a VM or natively). To install snapcraft on a Linux distro, 
first [install support for snaps](https://snapcraft.io/docs/installing-snapd), then install snapcraft as a snap with:

```bash
sudo snap install snapcraft
```

(note you will be promted to acknowledge you are installing a classic snap - use the `--classic` flag to acknowledge this)
#### Running snapcraft on MacOS

To install snapcraft on MacOS, see [this link](https://snapcraft.io/docs/installing-snapcraft#heading--macos). 
After doing so, follow the below instructions for "[Building with multipass](#building-with-multipass)".

#### Running snapcraft on Windows

To install snapcraft on Windows, you will need to run a Linux VM and follow the above instructions to install snapcraft as a snap. 
Note that if you are using WSL, only WSL2 with full Linux kernel support will work - you cannot use WSL with snapcraft and snaps. 
If you like, you can install multipass to launch a Linux VM if your Windows machine has Windows 10 Pro or Enterprise with Hyper-V support. 
See this [forum post](https://discourse.ubuntu.com/t/installing-multipass-for-windows/9547) for more details.

### Building with multipass

The easiest way to build the snap is using the multipass VM tool that snapcraft knows to use directly. 
After [installing multipass](https://multipass.run), just run 

```bash
snapcraft
```

### Building with LXD containers

Alternatively, you can instruct snapcraft to use LXD containers instead of multipass VM's. 
This requires installing LXD as documented [here](https://snapcraft.io/docs/build-on-lxd).

```bash
snapcraft --use-lxd
```

Note that if you are building on non-amd64 hardware, snapcraft won't be able to use it's default LXD container image, 
so you can follow the next section to create an LXD container to run snapcraft in destructive-mode natively in the container.

### Building inside external container/VM using native snapcraft

Finally, snapcraft can be run inside a VM, container or other similar build environment to build the snap without having snapcraft manage the environment 
(such as in a docker container where snaps are not available, or inside a VM launched from a build-farm without using nested VM's). 

This requires creating an Ubuntu 18.04 environment and running snapcraft (from the snap) inside the environment with `--destructive-mode`. 

#### LXD

Snaps run inside LXD containers just like they do outside the container, so all you need to do is launch an Ubuntu 20.04 container, 
install snapcraft and run snapcraft like follows:

```bash
$ lxc launch ubuntu:20.04 edgex
Creating edgex
Starting edgex
$ lxc exec edgex /bin/bash
root@edgex:~# sudo apt update && sudo apt install snapd squashfuse git -y
root@edgex:~# sudo snap install snapcraft --classic
root@edgex:~# git clone https://github.com/edgexfoundry/edgex-go
root@edgex:~# cd edgex-go && snapcraft --destructive-mode
```

#### Docker

Snapcraft is smart enough to detect when it is running inside a docker container specifically, 
to the point where no additional arguments are need to snapcraft when it is run inside the container. 
For example, the upstream snapcraft docker image can be used (only on x86_64 architectures unfortunately) like so:

```bash
docker run -it -v"$PWD":/build snapcore/snapcraft:stable bash -c "apt update && cd /build && snapcraft"
```

Note that if you are building your own docker image, you can't run snapd inside the container, and so to install snapcraft, 
the docker image must download the snapcraft snap and extract it as if it was installed normally inside `/snap` 
(same goes for the `core` and `core18` snaps). 
This is done by the Linux Foundation Jenkins server for the project's CI and you can see an example of that 
[here](https://github.com/edgexfoundry/ci-management/blob/master/shell/edgexfoundry-snapcraft.sh). 
The upstream docker image also does this, but only for x86_64 architectures.

#### Multipass / generic VM

To use multipass to create an Ubuntu 20.04 environment suitable for building the snap (i.e. when running natively on windows):

```bash
$ multipass launch focal -n edgex-snap-build
$ multipass shell edgex-snap-build
multipass@ubuntu:~$ git clone https://github.com/edgexfoundry/edgex-go
multipass@ubuntu:~$ cd edgex-go
multipass~ubuntu:~$ sudo snap install snapcraft --classic
multipass~ubuntu:~$ snapcraft --destructive-mode
```

The process should be similar for other VM's such as kvm, VirtualBox, etc. where you create the VM, clone the git repository, 
then install snapcraft as a snap and run with `--destructive-mode`. 

### Developing the snap

After building the snap from one of the above methods, you will have a binary snap package called `edgexfoundry_<latest version>_<arch>.snap`, 
which can be installed locally with the `--dangerous` flag:

```bash
sudo snap install edgexfoundry_<latest version>_<arch>.snap --dangerous
```

In addition, if you are using snapcraft with multipass VM's, you can speedup development by using `snapcraft try` and `snap try`:

1. Clean all the parts (optional to ensure a clean start)

```bash
snapcraft clean
```

2. Create a snap file with a prime folder placed in the root project directory

```bash
snapcraft prime --shell
```

It produces a prime directory instead of the snap file, and copies its prime directory to the current working directory on the host system - 
outside of any build environment container.

3. Install an unpacked snap using a bind mount to test out the snap defined in the project directory

```bash
sudo snap try prime --devmode
```

`snap try` works the same as `snap install`, but expects a directory. 

4. Change your code, then repeat step 2 and 3

5. Ship the tested snap file

```bash
sudo snap pack ./prime
```

#### Interfaces

After installing the snap, you will need to connect interfaces and restart the snap. 
Interfaces are automatically connected using snap store assertions when installing from the store, 
but when developing the snap and installing a revision locally, use the commands in this section to connect the interfaces.

To see which interfaces the snap is using, and which interfaces it could use but isn’t:

```bash
sudo snap connections edgexfoundry
```

```
Interface                         Plug                                  Slot                                                    Notes
content[edgex-secretstore-token]  edgexfoundry:edgex-secretstore-token  edgex-app-service-configurable:edgex-secretstore-token  -
content[edgex-secretstore-token]  edgexfoundry:edgex-secretstore-token  edgex-device-mqtt:edgex-secretstore-token               -
content[edgex-secretstore-token]  edgexfoundry:edgex-secretstore-token  edgex-device-snmp:edgex-secretstore-token               -
home                              edgexfoundry:home                     :home                                                   -
network                           edgexfoundry:network                  :network                                                -
network-bind                      edgexfoundry:network-bind             :network-bind                                           -
removable-media                   edgexfoundry:removable-media          -                                                       -
```

Manual connections:

```bash
snap connect <snap>:<plug interface> <snap>:<slot interface>
```

Here is an example:

```bash
sudo snap connect edgexfoundry:edgex-secretstore-token <your-snap>:<slot>
```

Please refer [here][secret-store-token] for further information.


[edgexfoundry]: https://snapcraft.io/edgexfoundry
[docs]: https://docs.edgexfoundry.org/2.2/getting-started/Ch-GettingStartedSnapUsers/#platform-snap
[secret-store-token]: https://docs.edgexfoundry.org/2.2/getting-started/Ch-GettingStartedSnapUsers/#secret-store-token
