# README #
Core Domain library for the Go implementation of EdgeX microservices.  This project contains the models that the other microservices use to pass around object data through http requests and to put object data into a database.

### What is this repository for? ###
* Domain objects for EdgeX microservices

### Installation ###
core-domain-go uses the glide vendoring tool - https://glide.sh/
To pull the dependecies into your workspace and build the project, make sure you are in the project directory and run:
```
go get github.com/edgexfoundry/core-domain-go
cd $GOPATH/src/github.com/edgexfoundry/core-domain-go
glide install
go install
```
