# EdgeX Foundry Support Logging Service
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Support Logging provides a centralized logging facility for all EdgeX microservices.  Logging service features a REST API for other micro services to add/query/delete logging requests. Two options of persistence--file or mongodb--are supported and are configurable.

# Install and Deploy Native #

### Prerequisites ###
Serveral EdgeX Foundry services depend on ZeroMQ for communications by default.  The easiest way to get and install ZeroMQ is to use or follow the following setup script:  https://gist.github.com/katopz/8b766a5cb0ca96c816658e9407e83d00.

**Note**: Setup of the ZeroMQ library is not supported on Windows plaforms.

### Installation and Execution ###
To fetch the code and build the microservice execute the following:

```
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# pull the 3rd party / vendor packages
make prepare
# build the microservice
make cmd/support-logging/support-logging
# get to the support logging microservice executable
cd cmd/support-logging
# run the microservice (may require other dependent services to run correctly)
./support-logging
```

# Install and Deploy via Docker Container #
This project has facilities to create and run Docker containers.  A Dockerfile is included in the repo. Make sure you have already run make prepare to update the dependecies. To do a Docker build using the included Docker file, run the following:

### Prerequisites ###
See https://docs.docker.com/install/ to learn how to obtain and install Docker.

### Installation and Execution ###

```
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# To create the Docker image
sudo make docker_support_logging
# To create a containter from the image
sudo docker create --name "[DOCKER_CONTAINER_NAME]" --network "[DOCKER_NETWORK]" [DOCKER_IMAGE_NAME]
# To run the container
sudo docker start [DOCKER_CONTAINER_NAME]
```

*Note* - creating and running the container above requires Docker network setup, may require dependent containers to be setup on that network, and appropriate port access configuration (among other start up parameters).  For this reason, EdgeX recommends use of Docker Compose for pulling, building, and running containers.  See The Getting Started Guides for more detail.
 

## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)

