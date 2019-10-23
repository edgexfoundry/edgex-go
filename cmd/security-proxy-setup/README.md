# EdgeX Foundry Security Service - Security Proxy Setup
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security-proxy-setup service.

## Features

- Reverse proxy for the existing edgex microservices
- Account creation with optional either OAuth2 or JWT authentication for existing services
- Account creation with arbitrary ACL gourp list

## Build

Use the Makefile in the root directory of the repository to build  security-proxy-setup:

```sh
 make cmd/security-proxy-setup/security-proxy-setup
```

This will create an executable located at `cmd/security-proxy-setup/` if successful.

## Run security-proxy-setup with different parameters

The binary supports multiple command line parameters 

```sh
 --init  //run init procedure for security service
 --insecureSkipVerify //skip server side SSL verification, mainly for self-signed cert
 --configfile //use different configuration file other than the default
 --reset //reset proxy by removing all customerization
 --useradd //user that needs to be added to consume the edgex services, needs to used with the 'group' parameter
 --group //group that the user belongs to
 --userdel //user that needs to be deleted in the the proxy services
```

An example of use of the parameters can be found from docker compose file

https://github.com/edgexfoundry/developer-scripts/master/releases/fuji/compose-files/docker-compose-fuji-1.1.0.yml

## Docker Build

Go to the root directory of the repository and use the Makefile to build the docker container image for `security-proxy-setup`:

```sh
make docker_security_proxy_setup
```

It should create a docker image with the name `edgexfoundry/docker_security_proxy_setup:<version>-dev` if sucessfully built.
