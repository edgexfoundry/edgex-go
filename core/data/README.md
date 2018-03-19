# README #
This repository is for the core data microservice for EdgeXFoundry written in the Go programming language.  The core data microservice is responsible for storing events and readings from device services and exposing these data through the export distro microservice.

### What is this repository for? ###
* Core data microservice for EdgeXFoundry

### Installation ###
This project uses Glide for dependency management - https://glide.sh/
To pull the dependencies and run the project, do the following:
```
go get github.com/edgexfoundry/core-data-go
cd $GOPATH/src/github.com/edgexfoundry/core-data-go
glide up
```
This project also uses ZeroMQ for sending messages to the export distro microservice.  If you use the Dockerfile to build and run a Docker image, then you do not have to worry about this dependecy as Docker handles it for you.  If you want to build the project locally, then you need to install ZeroMQ on your computer.
* Windows installers - http://zeromq.org/distro:microsoft-windows
* Get the software - http://zeromq.org/intro:get-the-software

Go to the installation directory for ZeroMQ and make sure there is a /bin folder and an /include folder.  There are 2 environment variables that you need to set for the build to suceed:
* CGO_LDFLAGS = "-L [PATH_TO_BIN_FOLDER]"
* CGO_CPPFLAGS = "-I [PATH_TO_INCLUDE_FOLDER]"

These variables will tell the CGO compiler where to find the C dependecies.

For convenience, you may want to check out this stackoverflow page for help:  https://stackoverflow.com/questions/41289619/how-to-install-zeromq-4-on-ubuntu-16-10-from-source

### Docker ###
This project can be built using Docker, and a Dockerfile is included in the repo.  Make sure you have already run 'glide up' to update the dependecies.  To build using the Docker file, run the following:
```
cd $GOPATH/src/github.com/edgexfoundry/core-data-go
docker build -t "[DOCKER_IMAGE_NAME]" .
```

To create a containter from the image run the following:
```
docker create --name "[DOCKER_CONTAINER_NAME]" --network "[DOCKER_NETWORK]" [DOCKER_IMAGE_NAME]
```

To run the container:
```
docker start [DOCKER_CONTAINER_NAME]
```
