# EdgeX Platform Snap
[![edgexfoundry](https://snapcraft.io/edgexfoundry/badge.svg)][edgexfoundry]

This directory contains the snap packaging of the EdgeX platform snap containing all reference core and security services, along with Support Scheduler and Support Notifications.

The platform snap is built automatically and published on the Snap Store as [edgexfoundry].

For usage instructions, please refer to Platform Snap section in [Getting Started using Snaps][docs].

## Limitations

See the [GitHub issues with snap label](https://github.com/edgexfoundry/edgex-go/issues?q=is%3Aopen+is%3Aissue+label%3Asnap) for current issues.

## Build from source

The snap is defined in [snapcraft.yaml](snapcraft.yaml) and built with [snapcraft](https://snapcraft.io/docs/snapcraft).

To build, execute the following command from the top-level directory of this repo:
```bash
snapcraft
```

This will create a snap package with `.snap` extension. It can be installed locally by setting the `--dangerous` flag:
```bash
sudo snap install --dangerous <snap-file>
```

Refer to [this guide](https://snapcraft.io/docs/iterating-over-a-build), for tips on how to quickly debug a build.

The [snapcraft overview](https://snapcraft.io/docs/snapcraft-overview) provides additional details.

### Interfaces
This snap has strict [confinement](https://snapcraft.io/docs/snap-confinement) which means that it runs in isolation up to a minimum level of access. The minimum access is granted by the connected snap [interfaces](https://snapcraft.io/docs/interface-management). Some of the interfaces such as [network](https://snapcraft.io/docs/network-interface) and [network-bind](https://snapcraft.io/docs/network-bind-interface) are connected automatically.

To see the available and connected interfaces for this snap:
```
$ snap connections edgexfoundry
Interface        Plug                                  Slot           Notes
content          edgexfoundry:edgex-secretstore-token  -              -
home             edgexfoundry:home                     :home          -
network          edgexfoundry:network                  :network       -
network-bind     edgexfoundry:network-bind             :network-bind  -
removable-media  edgexfoundry:removable-media          -              -
```

This shows five interface *plugs*, three of which are connected to corresponding system *slots*.

The `edgex-secretstore-token` snap plug makes it possible to send a token to locally installed EdgeX snaps, such as device and app service snaps. If both snaps are installed from the store and from the official provider, this connection would happen automatically.

A manual connection is possible by running:
```bash
sudo snap connect edgexfoundry:edgex-secretstore-token <consumer-snap>:edgex-secretstore-token
```

Please refer [here][secret-store-token] for further information.


[edgexfoundry]: https://snapcraft.io/edgexfoundry
[docs]: https://docs.edgexfoundry.org/3.0/getting-started/Ch-GettingStartedSnapUsers/#platform-snap
[secret-store-token]: https://docs.edgexfoundry.org/3.0/getting-started/Ch-GettingStartedSnapUsers/#secret-store-token
