# EdgeX Platform Snap
[![edgexfoundry](https://snapcraft.io/edgexfoundry/badge.svg)][edgexfoundry]

This directory contains the snap packaging of the EdgeX platform snap containing all reference core services along with several other security, supporting, application, and device services.

The platform snap is built automatically and published on the Snap Store as [edgexfoundry].

For usage instructions, please refer to App Service Configurable section in [Getting Started using Snaps][docs].

## Overview

This folder contains snap packaging for the EdgeX Foundry reference implementation.

The snap contains all of the EdgeX Go-based micro services from this repository, device-virtual, app-service-configurable
(for Kuiper integration), as well as all the necessary runtime components (e.g. Consul, Kong, Redis, ...) required to run an EdgeX instance.

The project maintains a rolling release of the snap on the `edge` channel that is rebuilt and published at least once daily.

The snap currently supports running on both `amd64` and `arm64` platforms.

## Installation

### Installing snapd
The snap can be installed on any system that supports snaps. 
You can see how to install snaps on your system [here](https://snapcraft.io/docs/installing-snapd).

However for full security confinement, the snap should be installed on an Ubuntu 18.04 LTS or later Desktop or Server, 
or a system running Ubuntu Core 18 or later.

### Installing EdgeX Foundry as a snap
The snap is published in the snap store at https://snapcraft.io/edgexfoundry.
You can see the current revisions available for your machine's architecture by running the command:

```bash
snap info edgexfoundry
```

The snap can be installed using `snap install`. To install the latest stable version:

```bash
sudo snap install edgexfoundry
```

To install the snap from the edge channel:

```bash
sudo snap install edgexfoundry --edge
```

**Note** - in general, installing from the edge channel is only recommended for development purposes. 
Depending on the state of the current development release, your mileage may vary.


You can also specify specific releases using the `--channel` option. For example to install the Jakarta (2.1) release of the snap:

```bash
sudo snap install edgexfoundry --channel=2.1
```

Lastly, on a system supporting it, the snap may be installed using GNOME (or Ubuntu) Software Center by searching for `edgexfoundry`.

**Note** - the snap has only been tested on Ubuntu Desktop/Server LTS releases (16.04 or later), as well as Ubuntu Core (16 or later).

**WARNING** - don't install the EdgeX snap on a system which is already running one of the included services 
(e.g. Consul, Redis, Vault, ...), as this may result in resource conflicts (i.e. ports) which could cause the `snap install` command to fail.

## Using the EdgeX snap

Upon installation, the following EdgeX services are automatically and immediately started:

* consul
* vault
* redis
* kong
* postgres
* core-data
* core-command
* core-metadata
* security-services (see [Security Services](#security-services) section below)

The following services are disabled by default:

* app-service-configurable (required for kuiper)
* device-virtual
* kuiper
* support-notifications
* support-scheduler
* sys-mgmt-agent

Any disabled services can be enabled and started up using `snap set`:

```bash
sudo snap set edgexfoundry support-notifications=on
```

To turn a service off (thereby disabling and immediately stopping it) set the service to off:

```bash
sudo snap set edgexfoundry support-notifications=off
```

All services which are installed on the system as systemd units, which if enabled will automatically start running when the system boots or reboots.

### Viewing logs
To view the logs for all services in the edgexfoundry snap use:

```bash
sudo snap logs edgexfoundry
```

Individual service logs may be viewed by specifying the service name:

```bash
sudo snap logs edgexfoundry.consul
```

Or by using the systemd unit name and `journalctl`:

```bash
journalctl -u snap.edgexfoundry.consul
```

### Configuring individual services

All default configuration files are shipped with the snap inside `$SNAP/config`, however because `$SNAP` isn't writable, 
all of the config files are copied during snap installation (specifically during the install hook, 
see `snap/hooks/install` in this repository) to `$SNAP_DATA/config`.

**Note** - `$SNAP` resolves to the path `/snap/edgexfoundry/current/` and `$SNAP_DATA` resolves to `/var/snap/edgexfoundry/current`.

The preferred way to change the configuration is to use [Configuration Overrides](#configuration-overrides) section below. 
It is also possible to change configuration directly via Consul's [UI](http://localhost:8500/ui/) or [kv REST API](https://www.consul.io/api/kv). 
Changes made to configuration in Consul require services to be restarted in order for the changes to take effect; 
the one exception are changes made to configuration items in a service's `[Writable]` section. 
Services that aren't started by default (see [Using the EdgeX snap](#using-the-edgex-snap) section above) 
*will* pickup any changes made to their config files when started.

Also it should be noted that use of Consul is enabled by default in the snap. It is not possible at this time to run the EdgeX services in
the snap with Consul disabled.


### Configuration Overrides
The EdgeX snap supports configuration overrides via its configure and install hooks which generate service-specific .env files 
which are used to provide a custom environment to the service, overriding the default configuration provided by the service's `configuration.toml` file. 
If a configuration override is made after a service has already started, then the service must be **restarted** via command-line 
(e.g. `snap restart edgexfoundry.<service>`), or [snapd's REST API](https://snapcraft.io/docs/snapd-api). 
If the overrides are provided via the snap configuration defaults capability of a gadget snap, 
the overrides will be picked up when the services are first started.

The following syntax is used to specify service-specific configuration overrides:

`env.<service>.<stanza>.<config option>`

For instance, to setup an override of Core Data's port use:

```bash
sudo snap set edgexfoundry env.core-data.service.port=2112
```

And restart the service:

```bash
sudo snap restart edgexfoundry.core-data
```

**Note** - at this time changes to configuration values in the [Writable] section are not supported.

For details on the mapping of configuration options to config options, 
please refer to [Service Environment Configuration Overrides](#service-environment-configuration-overrides) section.

### Security services

Currently, the security services are enabled by default. The security services constitute the following components:
* kong-daemon (API Gateway a.k.a. Reverse Proxy)
* postgres (kong's database)
* vault (Secret Store)

Oneshot services which perform the necessary security setup and stop:
* security-proxy-setup (kong setup)
* security-secretstore-setup (vault setup)
* security-bootstrapper-redis (secure redis setup)
* security-consul-bootstrapper (secure consul setup)

#### Secret Store
Vault is used by EdgeX for secret management (e.g. certificates, keys, passwords, ...) and is referred to as the Secret Store.

Use of Secret Store by all services can be disabled globally, but doing so will also disable the API Gateway, as it depends on the Secret Store.
Thus the following command will disable both:

```bash
sudo snap set edgexfoundry security-secret-store=off
```

*NOTE* - disabling security after the snap is installed is a convenience for developers. The snap will not allow the Secret Store to be
re-enabled. The only way to re-enable the Secret Store is to re-install the snap.

#### API Gateway
Kong is used for access control to the EdgeX services from external systems and is referred to as the API Gateway. 

For more details please refer to the EdgeX API Gateway [documentation](https://docs.edgexfoundry.org/2.1/security/Ch-APIGateway/).


The API Gateway can be disabled by using the following command:

```bash
sudo snap set edgexfoundry security-proxy=off
```

**Note** - by default all services in the snap except for the API Gateway are restricted to listening on 'localhost' 
(i.e. the services are not addressable from another system). In order to make a service accessible remotely, 
the appropriate configuration override of the 'Service.ServerBindAddr' needs to be made 
(e.g. ```sudo snap set edgexfoundry env.core-data.service.server-bind-addr=0.0.0.0```).

#### API Gateway user setup

##### JWT tokens

Before the API Gateway can be used, a user and group must be created and a JWT access token generated.

1. The first step is to create a public/private keypair for the new user, which can be done with

```bash
# Create private key:
openssl ecparam -genkey -name prime256v1 -noout -out private.pem

# Create public key:
openssl ec -in private.pem -pubout -out public.pem
```

2. The next step is to create the user. The easiest way to create a single API gateway user is to use `snap set` to set two values as follows:

```bash
# set user=username,user id,algorithm (ES256 or RS256)
sudo snap set edgexfoundry env.security-proxy.user=user01,USER_ID,ES256

# set public-key to the contents of a PEM-encoded public key file
sudo snap set edgexfoundry env.security-proxy.public-key="$(cat public.pem)"
```

To create multiple users, use the secrets-config command. You need to provide the following:

- The username
- The public key
- The API Gateway Admin JWT token
- (optionally) ID. This is a unique string identifying the credential. It will be required in the next step to
create the JWT token. If you don't specify it,
then an autogenerated one will be output by the secrets-config command
```bash

# get API Gateway/Kong token
JWT_FILE=/var/snap/edgexfoundry/current/secrets/security-proxy-setup/kong-admin-jwt
JWT=`sudo cat ${JWT_FILE}`

# use secrets-config to add user
edgexfoundry.secrets-config proxy adduser --token-type jwt --user user01 --algorithm ES256 --public_key public.pem --id USER_ID --jwt ${JWT}
```

3. Finally, you need to generate a token using the user ID which you specified:

```bash
# get token
TOKEN=`edgexfoundry.secrets-config proxy jwt --algorithm ES256 --private_key private.pem --id USER_ID --expiration=1h`

# Keep this token in a safe place for future reuse as the same token cannot be regenerated or recovered using the secret-config CLI
echo $TOKEN
```

Alternatively , you can generate the token on a different device using a bash script:

```bash
header='{
    "alg": "ES256",
    "typ": "JWT"
}'

TTL=$((EPOCHSECONDS+3600)) 

payload='{
    "iss":"USER_ID",
    "iat":'$EPOCHSECONDS', 
    "nbf":'$EPOCHSECONDS',
    "exp":'$TTL' 
}'

JWT_HEADER=`echo -n $header | openssl base64 -e -A | sed s/\+/-/ | sed -E s/=+$//`
JWT_PAYLOAD=`echo -n $payload | openssl base64 -e -A | sed s/\+/-/ | sed -E s/=+$//`
JWT_SIGNATURE=`echo -n "$JWT_HEADER.$JWT_PAYLOAD" | openssl dgst -sha256 -binary -sign private.pem  | openssl asn1parse -inform DER  -offset 2 | grep -o "[0-9A-F]\+$" | tr -d '\n' | xxd -r -p | base64 -w0 | tr -d '=' | tr '+/' '-_'`
TOKEN=$JWT_HEADER.$JWT_PAYLOAD.$JWT_SIGNATURE
```

4. Once you have the token you can access the API Gateway as follows:

The JWT token must be included
via an HTTP `Authorization: Bearer <access-token>` header on any REST calls used to access EdgeX services via the API Gateway. 

Example:

```bash
curl -k -X GET https://localhost:8443/core-data/api/v2/ping? -H "Authorization: Bearer $TOKEN"
```


#### API Gateway TLS certificate setup

By default Kong is configured with a self-signed TLS certificate (which you find in `/var/snap/edgexfoundry/current/kong/ssl/kong-default-ecdsa.crt`). 
It is also possible to install your own TLS certificate to be used by the gateway. The steps to do so are as follows:

1. Start by provisioning a TLS certificate to use. You can use a number of tools for that, such as `openssl` or the `edgeca` snap:

```bash
sudo snap install edgeca
edgeca gencsr --cn localhost --csr csrfile --key csrkeyfile
edgeca gencert -o localhost.cert -i csrfile -k localhost.key
```

2. Then install the certificate:

```bash
sudo snap set edgexfoundry env.security-proxy.tls-certificate="$(cat localhost.cert)"
sudo snap set edgexfoundry env.security-proxy.tls-private-key="$(cat localhost.key)"
```

3. Specify the EdgeCA Root CA certificate with `--cacert` for validation of the new certificate:

```bash
curl -v --cacert /var/snap/edgeca/current/CA.pem -X GET https://localhost:8443/core-data/api/v2/ping? -H "Authorization: Bearer $TOKEN"
```

Optionally, to specify a server name other than `localhost`, set the `tls-sni` configuration setting first. Example:

```bash
# generate certificate and private key
edgeca gencsr --cn server01 --csr csrfile --key csrkeyfile
edgeca gencert -o server.cert -i csrfile -k server.key

# To set the certificate again, you first need to clear the current values by setting them to an empty string:
sudo snap set edgexfoundry env.security-proxy.tls-certificate=""
sudo snap set edgexfoundry env.security-proxy.tls-private-key=""

# set tls-sni
sudo snap set edgexfoundry env.security-proxy.tls-sni="server01"

# and then provide the certificate and key
sudo snap set edgexfoundry env.security-proxy.tls-certificate="$(cat server.cert)"
sudo snap set edgexfoundry env.security-proxy.tls-private-key="$(cat server.key)"

# connect
curl -v --cacert /var/snap/edgeca/current/CA.pem -X GET https://server01:8443/core-data/api/v2/ping? -H "Authorization: Bearer $TOKEN"
```

## Limitations

[See the GitHub issues with label snap for current issues.](https://github.com/edgexfoundry/edgex-go/issues?q=is%3Aopen+is%3Aissue+label%3Asnap)

## Service environment configuration overrides
**Note** - all of the configuration options below must be specified with the prefix: `env.<service>.` where `<service>` is one of the following:

  - core-data
  - core-metadata
  - core-command
  - support-notifications
  - support-scheduler
  - app-service-configurable
  - device-virtual
  - security-secret-store
  - security-proxy
  - sys-mgmt-agent

Example: `snap set edgexfoundry env.device-virtual.service.port=7777`


| [Service]                     | Description                                                  |
| ----------------------------- | ------------------------------------------------------------ |
| service.health-check-interval | Sets the interval for Registry heal check callback           |
| service.host:                 | Sets the hostname or IP address of the service.              |
| service.port                  | Sets the HTTP port of the service.                           |
| service.server-bind-addr      | Specifies an IP address or hostname  for ListenAndServe to bind to, such as 0.0.0.0 |
| service.startup-msg           | Specifies a string to log once service  initialization and startup is completed. |
| service.max-result-count      | Specifies the maximum size list supported in response to REST calls to other services. |
| service.max-request-size      | Defines the maximum size of http request body in bytes       |
| service.request-timeout       | Specifies a timeout (in milliseconds) for processing REST request calls from other services. |

| [SecretStore]                           | Description                                                  |
| --------------------------------------- | ------------------------------------------------------------ |
| secret-store.secrets-file               | Sets the path of the JSON file containing secrets needed for seeding. If left blank, seeding is skipped. |
| secret-store.disable-scrub-secrets-file | If set, `secret-store.secrets-file` will not overwrite the default secret. |


| [Clients]                          | Description                          |
| ---------------------------------- | ------------------------------------ |
| clients.core-command.port          | Clients' core-command port.          |
| clients.core-data.port             | Clients' core-data port port.        |
| clients.core-metadata.port         | Clients' core-metadata port.         |
| clients.support-notifications.port | Clients' support-notifications port. |
| clients.support-scheduler.port     | Clients' support-scheduler port.     |


| [MessageQueue] (core-data only)   | Description                                                  |
| --------------------------------- | ------------------------------------------------------------ |
| messagequeue.type                 | Indicates the message bus implementation to use, i.e. mqtt, redis streams. |
| messagequeue.protocol             | Indicates the protocol to use when accessing the message bus. |
| messagequeue.host                 | Sets the hostname or IP address of the broker, if applicable. |
| messagequeue.port                 | Defines the port on which to access the message bus.         |
| messagequeue.publish-topic-prefix | Indicates the topic prefix the data is published to.         |
| messagequeue.subscribe-topic      | Indicates the topic in which to subscribe.                   |
| messagequeue.auth-mode            | Specifies the type of secure connection to the message bus which are 'none', 'usernamepassword','clientcert' or 'cacert'. Not all option supported by each implementation. ZMQ doesn't support any Authmode beyond 'none', RedisStreams only supports 'none' & 'usernamepassword' while MQTT supports all options. |
| messagequeue.secret-name          | Sets the name of the secret in the SecretStore that contains the Auth Credentials. The credential are dynamically loaded using this name and store the Option property below where the implementation expected to find them. |
| messagequeue.subscribe-enabled    | Indicates whether enable the subscription to the Message Queue |

| [Trigger]                                                 | Description                                                  |
| --------------------------------------------------------- | ------------------------------------------------------------ |
| trigger.edgex-message-bus.subscribe-host.port             | Defines the port on which to access the Edgex Message Bus's subscribe host. |
| trigger.edgex-message-bus.subscribe-host.protocol         | Indicates the protocol to use when accessing the Edgex Message Bus's subscribe host. |
| trigger.edgex-message-bus.subscribe-host.subscribe-topics | Indicates the topic in which Edgex Message Bus's subscribe host to subscribe. |
| trigger.edgex-message-bus.publish-host.port               | Defines the port on which to access the Edgex Message Bus's publish host. |
| trigger.edgex-message-bus.publish-host.protocol           | Indicates the protocol to use when accessing the Edgex Message Bus's publish host. |
| trigger.edgex-message-bus.publish-host.publish-topic      | Indicates the topic in which Edgex Message Bus's publish host to subscribe. |

### API Gateway settings (prefix: `env.security-proxy.`)

| API Gateway setting | Description                                                  |
| ------------------- | ------------------------------------------------------------ |
| add-proxy-route     | The add-proxy-route setting is a comma-separated list of paired additional service names and URLs to be added to the API Gateway (aka Kong). See [documentation](https://docs.edgexfoundry.org/2.1/security/Ch-APIGateway/#configuration-of-adding-microservices-routes-for-api-gateway) NOTE - this setting is not a configuration override, it's a top-level environment variable used by the security-proxy-setup. |

### Secret Store settings (prefix: `env.security-secret-store.`)

| API Gateway setting    | Description                                                  |
| ---------------------- | ------------------------------------------------------------ |
| add-secretstore-tokens | Defines a comma-separated list of service keys to be added to the list of Vault tokens that security-file-token-provider (launched by security-secretstore-setup) creates. It is set to a default list of additional services by the snap, so be sure to examine the default setting before providing a custom list of services. NOTE - this setting is not a configuration override, it's a top-level environment variable used by the security-secretstore-setup. |
| add-known-secrets      | The add-known-secrets setting is a comma-separated list of secret paths and associated services. It's used to provision the specified secret for the given service in Vault. It is set to a default list of additional services by the snap, so be sure to examine the default setting before providing a custom list of services. NOTE - this setting is not a configuration override, it's a top-level environment variable used by the security-secretstore-setup. |
| default-token-ttl      | The default-token-ttl setting is a Go Duration string, a sequence of decimal numbers, each with optional fraction and a unit suffix (e.g. "ns", "us" (or "µs"), "ms", "s", "m", "h"). It's used to set the TTL of vault tokens generated for EdgeX services during bootstrap. This setting can be used to increase (or decrease) the default TTL (one hour). If the TTL of a token expires before a service is started, the service will not be able to access the Secret Store. |

### Security bootstrapper settings (prefix: `env.security-bootstrapper.`)

| setting                | Description                                                  |
| ---------------------- | ------------------------------------------------------------ |
| add-registry-acl-roles | A comma separated list of registry role names to be added. |


### Support Notifications settings (prefix: `env.support-notifications.`)

| [Smtp]                       | Description                                                  |
| ---------------------------- | ------------------------------------------------------------ |
| smtp.host                    | SMTP hostname                                                |
| smtp.username                | SMTP username                                                |
| smtp.password                | SMTP password                                                |
| smtp.port                    | SMTP port                                                    |
| smtp.sender                  | Notification message sender                                  |
| smtp.enable-self-signed-cert | Enable support for invalid (self-signed) certificates        |
| smtp.subject                 | Notification message subject                                 |
| smtp.auth-mode               | User need to store the credential via the /secret API before sending the email notification. AuthMode is the SMTP authentication mechanism. Currently, `usernamepassword' is the only AuthMode supported by this service, and the secret keys are 'username' and 'password'. |

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