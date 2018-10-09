# EdgeX Foundry Services

[![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/edgex-go)](https://goreportcard.com/report/github.com/edgexfoundry/edgex-go)
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

EdgeX Foundry is a vendor-neutral open source project hosted by The Linux Foundation building a common open framework for IoT edge computing. At the heart of the project is an interoperability framework hosted within a full hardware- and OS-agnostic reference software platform to enable an ecosystem of plug-and-play components that unifies the marketplace and accelerates the deployment of IoT solutions. This repository contains the Go implementation of EdgeX Foundry microservices. It also includes files for building the services, containerizing the services, and initializing (bootstrapping) the services.

## Install and Deploy Native

## Prerequisites

### pkg-config

`go get github.com/rjeczalik/pkgconfig/cmd/pkg-config`

### ZeroMQ

Several EdgeX Foundry services depend on ZeroMQ for communications by default.

The easiest way to get and install ZeroMQ on Linux is to use or follow the following setup script: https://gist.github.com/katopz/8b766a5cb0ca96c816658e9407e83d00.

For MacOS, use brew: `brew install zeromq`. Please note that the necessary `pc` file will need to be added to the `PKG_CONFIG_PATH` environment variable. For example `PKG_CONFIG_PATH=/usr/local/Cellar/zeromq/4.2.5/lib/pkgconfig/`

**Note**: Setup of the ZeroMQ library is not supported on Windows plaforms.

## Configuration

This section describes the configuration needed for various services.

### Core Services

Configuration of Core Services is via `.toml` files for each respective service.

| Service       | Configuration                            |
| ------------- | ---------------------------------------- |
| Core Data     | `cmd/core-data/res/configuration.toml`     |
| Core Metadata | `cmd/core-metadata/res/configuration.toml` |

#### Database

Configure `[Databases.Primary]` section of the respective `.toml` file. For example to use Redis

```TOML
[Databases]
  [Databases.Primary]
  Host = 'localhost'
  Name = 'exportclient'
  Port = 6379
  Type = 'redisdb'
```

### Export Client

Configuration of Export Client is via `cmd/export-client/res/configuration.tml`

#### Database

Configure `[Databases.Primary]` section of the respective `.toml` file. For example to use Redis

```TOML
[Databases]
  [Databases.Primary]
  Host = 'localhost'
  Name = 'exportclient'
  Port = 6379
  Type = 'redisdb'
```

## Installation and Execution

To fetch the code and build the microservice execute the following:

```sh
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# pull the 3rd party / vendor packages
make prepare
# build the microservices
make build
# run the services
make run
```

**Note** You will need to have the database and Consul running before you execute `make run`. If you don't want to install them, start them via their respective Docker containers.

## Install and Deploy via Docker Container

This project has facilities to create and run Docker containers.

## Prerequisites

See https://docs.docker.com/install/ to learn how to obtain and install Docker.

## Installation and Execution

```sh
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# To create the Docker images
sudo make docker
# To run the containers
sudo make run_docker
```

## Install and Deploy via Snap Package

EdgeX Foundry is also available as a snap package, for more details
on the snap, including how to install it, please refer to [EdgeX snap](https://github.com/edgexfoundry/edgex-go/blob/master/snap/README.md)

## Docker Hub

EdgeX images are kept on organization's [DockerHub page](https://hub.docker.com/u/edgexfoundry/).
They can be run in orchestration via official [docker-compose.yml](https://github.com/edgexfoundry/developer-scripts/blob/master/compose-files/docker-compose.yml).

The simplest way is to do this via prepared script in `bin` directory:

```sh
cd bin
./edgex-docker-launch.sh
```

### Additional Options

`make run_docker` as described earlier invokes this script.

If you want to use your own Docker compose file, create `local-docker-compose.yml` and override the script defaults by setting the environment variable `EDGEX_COMPOSE_FILE` to the full path of your compose file.

The environment variable `EDGEX_SERVICES` can be used to override the services started by the script.

The environment variable `EDGEX_CORE_DB` can be set to the DB service to override the default DB in the script.

For example, to start all the services except metadata data and export-client

```sh
cd bin
EDGEX_SERVICES="logging command export-distro notifications" EDGEX_CORE_DB=redis EDGEX_COMPOSE_FILE=../docker/local-docker-compose.yml bin/edgex-docker-launch.sh
```

## Compiled Binaries

During development phase, it is important to run compiled binaries (not containers).

There is a script in `bin` directory that can help you launch the whole EdgeX system:

```sh
cd bin
./edgex-launch.sh
```

## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)
