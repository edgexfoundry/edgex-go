# EdgeX Foundry Services
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/edgex-go/job/main/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/edgex-go/job/main/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/edgex-go/branch/main/graph/badge.svg?token=Y3mpessZqk)](https://codecov.io/gh/edgexfoundry/edgex-go) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/edgex-go)](https://goreportcard.com/report/github.com/edgexfoundry/edgex-go) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/edgex-go?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/edgex-go/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/edgex-go?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/edgex-go)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/edgex-go) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/edgex-go)](https://github.com/edgexfoundry/edgex-go/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/edgex-go)](https://github.com/edgexfoundry/edgex-go/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/edgex-go-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/edgex-go)](https://github.com/edgexfoundry/edgex-go/commits)

> **Warning**  
> The **main** branch of this repository contains work-in-progress development code for the upcoming release, and is **not guaranteed to be stable or working**.
> It is only compatible with the [main branch of edgex-compose](https://github.com/edgexfoundry/edgex-compose) which uses the Docker images built from the **main** branch of this repo and other repos.
>
> **The source for the latest release can be found at [Releases](https://github.com/edgexfoundry/edgex-go/releases).**

EdgeX Foundry is a vendor-neutral open source project hosted by The Linux Foundation building a common open framework for IoT edge computing.  At the heart of the project is an interoperability framework hosted within a full hardware- and OS-agnostic reference software platform to enable an ecosystem of plug-and-play components that unifies the marketplace and accelerates the deployment of IoT solutions.  This repository contains the Go implementation of EdgeX Foundry microservices.  It also includes files for building the services, containerizing the services, and initializing (bootstrapping) the services.

## Build with NATS Messaging
Currently, the NATS Messaging capability (NATS MessageBus) is opt-in at build time. This means that the published Docker images and Snaps do not include the NATS messaging capability.

The following make commands will build the local binaries or local Docker images with NATS messaging capability included for the Core and Support services.

```makefile
make build-nats
make docker-nats
```

The locally built Docker images can then be used in place of the published Docker images in your compose file.
See [Compose Builder](https://github.com/edgexfoundry/edgex-compose/tree/main/compose-builder#gen) `nat-bus` option to generate compose file for NATS and local dev images.

## Get Started

EdgeX provides docker images in our organization's [DockerHub page](https://hub.docker.com/u/edgexfoundry/).
They can be launched easily with **docker-compose**.

The simplest way to get started is to fetch the latest docker-compose.yml and start the EdgeX containers:

```sh
release="main" # or "jakarta" for latest
wget https://raw.githubusercontent.com/edgexfoundry/edgex-compose/${release}/docker-compose.yml
docker-compose up -d
```

You can check the status of your running EdgeX services by going to http://localhost:8500/

Now that you have EdgeX up and running, you can follow our [API Walkthrough](https://docs.edgexfoundry.org/2.1/walk-through/Ch-Walkthrough/) to learn how the different services work together to connect IoT devices to cloud services.

## Running EdgeX with security components

Starting with the Fuji release, EdgeX includes enhanced security features that are enabled by default. There are a few major components that are responsible for security
features: 

- Security-secretstore-setup
- Security-proxy-setup

As part of Ireland release, the `security-secrets-setup` service is no more as internal service-to-service communication will not run in TLS by default in a single box.

When security features are enabled, additional steps are required to access the resources of EdgeX.

1. The user needs to create an access token and associate every REST request with the access token. 
2. The exported external ports (such as 59880, 59881 etc.) will be inaccessible for security purposes. Instead, all REST requests need to go through the proxy. The proxy will redirect the request to the individual microservices on behalf of the user.

Sample steps to create an access token and use the token to access EdgeX resources can be found here: [Security Components](SECURITY.md)

## Other installation and deployment options

### Snap Package

The components in this repository are available as a snap package.
For more details on the snap, including how to build and install it, please refer to the [snap](snap) directory.

### Native binaries

#### Prerequisites

##### Go

- The current targeted version of the Go language runtime for release artifacts is v1.18.x
- The minimum supported version of the Go language runtime is v1.18.x

#### Installation and Execution

EdgeX is organized as Go Modules; there is no requirement to set the GOPATH or
GO111MODULE envrionment variables nor is there a requirement to root all the components under ~/go
(or $GOPATH) and use the `go get` command. In other words,

```sh
git clone git@github.com:edgexfoundry/edgex-go.git
cd edgex-go
make build
```

If you do want to root everthing under $GOPATH, you're free to use that pattern as well

```sh
GO111MODULE=on && export GO111MODULE
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
make build
```

#### Deploy EdgeX

Recommended deployment of EdgeX services is either Docker or Snap. See [Getting Started with Docker](https://docs.edgexfoundry.org/2.0/getting-started/Ch-GettingStartedUsers/) or [Getting Started with Snap](https://docs.edgexfoundry.org/2.0/getting-started/Ch-GettingStartedSnapUsers/) for more details. 

#### Hybrid for debug/testing

If you need to run and/or debug one of the services locally, simply stop the docker container running that service and run the service locally from command-line or from your debugger. All executables are located in the `cmd/<service-name>` folders. See [Working in a Hybrid Environment](https://docs.edgexfoundry.org/2.0/getting-started/Ch-GettingStartedHybrid/) for more details.

> *Note that this works best when running the service in non-secure mode. i.e. with environment variable `EDGEX_SECURITY_SECRET_STORE=false`*

### Build your own Docker Containers

In addition to running the services directly, Docker and Docker Compose can be used.

#### Prerequisites

See [the install instructions](https://docs.docker.com/install/) to learn how to obtain and install Docker.

#### Build

Follow the "Installation and Execution" steps above for obtaining and building the code, then

```sh
make docker 
```

#### Delayed Start Go Builds For Developers

Currently for EdgeX core services except support services (support-notification and support-scheduler services), the delayed start feature from the dependency go-mod-bootstrap / go-mod-secrets modules are excluded in go builds by default.
If you want to **include** the delayed start feature in the builds for these services, please change the [Makefile in this directory](Makefile). In particular, change the following boolean flag from `false` to `true` before the whole docker builds.

```text
INCLUDE_DELAYED_START_BUILD_CORE:="false"
```

For support services, the delayed start feature is included by default as the default behavior of them are not started right away in Snap. Similarly, you can change the default and **exclude** it by modifying the boolean flag from `true` to `false` in the Makefile:

```text
INCLUDE_DELAYED_START_BUILD_SUPPORT:="true"
```

#### Run 

The **Compose Builder** tool has the `dev` option to generate and run EdgeX compose files using locally built images for above. See [Compose Builder README](https://github.com/edgexfoundry/edgex-compose/tree/main/compose-builder#readme) for more details.

```bash
make run no-secty dev
```

> *Note that this run all the edgex-go services using the locally built images.*

#### Community

- Discussion: https://github.com/orgs/edgexfoundry/discussions
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo

## License

[Apache-2.0](LICENSE)

## Versioning

Please refer to the EdgeX Foundry [versioning policy](https://wiki.edgexfoundry.org/pages/viewpage.action?pageId=21823969) for information on how EdgeX services are released and how EdgeX services are compatible with one another.  Specifically, device services (and the associated SDK), application services (and the associated app functions SDK), and client tools (like the EdgeX CLI and UI) can have independent minor releases, but these services must be compatible with the latest major release of EdgeX.

## Long Term Support

Please refer to the EdgeX Foundry [LTS policy](https://wiki.edgexfoundry.org/pages/viewpage.action?pageId=69173332) for information on support of EdgeX releases. The EdgeX community does not offer support on any non-LTS release outside of the latest release.
