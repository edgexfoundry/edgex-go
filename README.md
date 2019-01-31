# EdgeX Foundry Services
[![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/edgex-go)](https://goreportcard.com/report/github.com/edgexfoundry/edgex-go)
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

EdgeX Foundry is a vendor-neutral open source project hosted by The Linux Foundation building a common open framework for IoT edge computing.  At the heart of the project is an interoperability framework hosted within a full hardware- and OS-agnostic reference software platform to enable an ecosystem of plug-and-play components that unifies the marketplace and accelerates the deployment of IoT solutions.  This repository contains the Go implementation of EdgeX Foundry microservices.  It also includes files for building the services, containerizing the services, and initializing (bootstrapping) the services.

# Get Started
EdgeX provides docker images in our organization's [DockerHub page](https://hub.docker.com/u/edgexfoundry/).
They can be launched easily with **docker-compose**.

The simplest way to get started is to fetch the latest docker-compose.yml and start the EdgeX containers:
```
wget https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/compose-files/docker-compose.yml
docker-compose up -d
```
You can check the status of your running EdgeX services by going to http://localhost:8500/

Now that you have EdgeX up and running, you can follow our [API Walkthrough](https://docs.edgexfoundry.org/Ch-Walkthrough.html) to learn how the different services work together to connect IoT devices to cloud services.

# Other installation and deployment options
## Snap Package
EdgeX Foundry is also available as a snap package, for more details
on the snap, including how to install it, please refer to [EdgeX snap](https://github.com/edgexfoundry/edgex-go/blob/master/snap/README.md)


## Native binaries

### Prerequisites
#### Glide
`curl https://glide.sh/get | sh`

#### pkg-config
`go get github.com/rjeczalik/pkgconfig/cmd/pkg-config`

#### ZeroMQ
Several EdgeX Foundry services depend on ZeroMQ for communications by default.

The easiest way to get and install ZeroMQ on Linux is to use this [setup script](https://gist.github.com/katopz/8b766a5cb0ca96c816658e9407e83d00).

For macOS, use brew: 
```
brew install zeromq
``` 

For directions installing ZeroMQ on Windows, please see [the Windows documentation.](ZMQWindows.md)

#### pkg-config

The necessary file will need to be added to the `PKG_CONFIG_PATH` environment variable.
 
On Linux, add this line to your local profile:
```bash
export PKG_CONFIG_PATH=/usr/local/Cellar/zeromq/4.2.5/lib/pkgconfig/
```

For macOS, install the package with brew:
```bash
brew install pkg-config
```

### Installation and Execution
To fetch the code and build the microservice execute the following:

```
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

**Note** You will need to have the database running before you execute `make run`. If you don't want to install a database locally, you can bring one up via their respective Docker containers.

#### Compiled Binaries
During development phase, it is important to run compiled binaries (not containers).

There is a script in `bin` directory that can help you launch the whole EdgeX system:
```
cd bin
./edgex-launch.sh
```

## Build your own Docker Containers
This project has facilities to create and run your own Docker containers.

#### Prerequisites
See https://docs.docker.com/install/ to learn how to obtain and install Docker.

#### Installation and Execution

```
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# To remove any old build artifacts
make clean
# To check and satisfy build dependencies
make prepare
# To create the Docker images
sudo make docker
# To run the containers
sudo make run_docker
```

# Community
- Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/enQtNDgyODM5ODUyODY0LWVhY2VmOTcyOWY2NjZhOWJjOGI1YzQ2NzYzZmIxYzAzN2IzYzY0NTVmMWZhZjNkMjVmODNiZGZmYTkzZDE3MTA)
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo

# License
[Apache-2.0](LICENSE)
