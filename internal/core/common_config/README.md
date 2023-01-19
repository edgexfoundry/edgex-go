# EdgeX Foundry Core Common Config Bootstrapper Service
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Core Common Config Bootstrapper pushes the common settings into the Configuration Provider on start-up and provides the override flag common to all services to force re-pushing the configuration into the Configuration Provider.

# Install and Deploy Native #

### Installation and Execution ###
To fetch the code and build the microservice execute the following:

```
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# pull the 3rd party / vendor packages
make prepare
# build the microservice
make common_config
# get to the command microservice executable
cd cmd/core-common-config-bootstrapper
# run the microservice (may require other dependent services to run correctly) and provide the configuration provider using -cp
./core-common-config-bootstrapper -cp
```

# Install and Deploy via Docker Container #
This project has facilities to create and run Docker containers.  A Dockerfile is included in the repo. Make sure you have already run make prepare to update the dependencies. To do a Docker build using the included Docker file, run the following:

### Prerequisites ###
See https://docs.docker.com/install/ to learn how to obtain and install Docker.

### Installation and Execution ###

```
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# To create the Docker image
sudo make docker_core_common_config
# To create a containter from the image
sudo docker create --name "[DOCKER_CONTAINER_NAME]" --network "[DOCKER_NETWORK]" [DOCKER_IMAGE_NAME]
# To run the container
sudo docker start [DOCKER_CONTAINER_NAME]
```

*Note* - creating and running the container above requires Docker network setup, may require dependent containers to be setup on that network, and appropriate port access configuration (among other start up parameters).  For this reason, EdgeX recommends use of Docker Compose for pulling, building, and running containers.  See The Getting Started Guides for more detail.
 

## Community
- Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/enQtNDgyODM5ODUyODY0LWVhY2VmOTcyOWY2NjZhOWJjOGI1YzQ2NzYzZmIxYzAzN2IzYzY0NTVmMWZhZjNkMjVmODNiZGZmYTkzZDE3MTA)
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)

