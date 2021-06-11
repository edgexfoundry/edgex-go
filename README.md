# EdgeX Foundry Services
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/edgex-go/job/master/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/edgex-go/job/master/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/edgex-go/branch/master/graph/badge.svg?token=Y3mpessZqk)](https://codecov.io/gh/edgexfoundry/edgex-go) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/edgex-go)](https://goreportcard.com/report/github.com/edgexfoundry/edgex-go) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/edgex-go?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/edgex-go/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/edgex-go?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/edgex-go)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/edgex-go) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/edgex-go)](https://github.com/edgexfoundry/edgex-go/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/edgex-go)](https://github.com/edgexfoundry/edgex-go/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/edgex-go-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/edgex-go)](https://github.com/edgexfoundry/edgex-go/commits)


EdgeX Foundry is a vendor-neutral open source project hosted by The Linux Foundation building a common open framework for IoT edge computing.  At the heart of the project is an interoperability framework hosted within a full hardware- and OS-agnostic reference software platform to enable an ecosystem of plug-and-play components that unifies the marketplace and accelerates the deployment of IoT solutions.  This repository contains the Go implementation of EdgeX Foundry microservices.  It also includes files for building the services, containerizing the services, and initializing (bootstrapping) the services.

## Get Started

EdgeX provides docker images in our organization's [DockerHub page](https://hub.docker.com/u/edgexfoundry/).
They can be launched easily with **docker-compose**.

The simplest way to get started is to fetch the latest docker-compose.yml and start the EdgeX containers:

```sh
release="pre-release" # or "hanoi" for latest
wget -O docker-compose.yml \
https://raw.githubusercontent.com/edgexfoundry/edgex-compose/master/releases/${release}/docker-compose-${release}.yml
docker-compose up -d
```

You can check the status of your running EdgeX services by going to http://localhost:8500/

Now that you have EdgeX up and running, you can follow our [API Walkthrough](https://docs.edgexfoundry.org/1.3/walk-through/Ch-Walkthrough/) to learn how the different services work together to connect IoT devices to cloud services.

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

EdgeX Foundry is also available as a snap package, for more details
on the snap, including how to install it, please refer to [EdgeX snap](https://github.com/edgexfoundry/edgex-go/blob/master/snap/README.md)

### Native binaries

#### Prerequisites

##### Go

- The current targeted version of the Go language runtime for release artifacts is v1.16.x
- The minimum supported version of the Go language runtime is v1.16.x

##### pkg-config

`go get github.com/rjeczalik/pkgconfig/cmd/pkg-config`

##### ZeroMQ

Several EdgeX Foundry services depend on ZeroMQ for communications by default.

The easiest way to get and install ZeroMQ on Linux is to use this [setup script](https://gist.github.com/katopz/8b766a5cb0ca96c816658e9407e83d00).

For macOS, use brew:

```sh
brew install zeromq
```

For directions installing ZeroMQ on Windows, please see [the Windows documentation.](ZMQWindows.md)

##### pkg-config

The necessary file will need to be added to the `PKG_CONFIG_PATH` environment variable.

On Linux, add this line to your local profile:

```sh
export PKG_CONFIG_PATH=/usr/local/Cellar/zeromq/4.2.5/lib/pkgconfig/
```

For macOS, install the package with brew:

```sh
brew install pkg-config
```

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

To start EdgeX

```sh
make run
```

or

```sh
cd bin
./edge-launch.sh
```

**Note** You must have a database (Redis) running before the services will operate
correctly. If you don't want to install a database locally, you can host one via Docker. You may
also need to change the `configuration.toml` files for one or more of the services.

### Build your own Docker Containers

In addition to running the services directly, Docker and Docker Compose can be used.

#### Prerequisites

See [the install instructions](https://docs.docker.com/install/) to learn how to obtain and install Docker.

#### Installation and Execution

Follow the "Installation and Execution" steps above for obtaining and building the code, then

```sh
make docker run_docker
```

**Note** The default behavior is to use Redis for the database.

## Community

- Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/zt-65tadqv4-DtLS6eiz_AC7_FeJSdSm8A)
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo

## License

[Apache-2.0](LICENSE)

## Versioning

Please refer to the EdgeX Foundry [versioning policy](https://wiki.edgexfoundry.org/pages/viewpage.action?pageId=21823969) for information on how EdgeX services are released and how EdgeX services are compatible with one another.  Specifically, device services (and the associated SDK), application services (and the associated app functions SDK), and client tools (like the EdgeX CLI and UI) can have independent minor releases, but these services must be compatible with the latest major release of EdgeX.

## Long Term Support

Please refer to the EdgeX Foundry [LTS policy](https://wiki.edgexfoundry.org/display/FA/Long+Term+Support) for information on support of EdgeX releases. The EdgeX community does not offer support on any non-LTS release outside of the latest release.
