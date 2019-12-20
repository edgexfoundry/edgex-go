# EdgeX Foundry Security Service - Security Proxy Setup
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security-proxy-setup service.

## Features

- Reverse proxy for EdgeX microservices
- Account creation with optional either OAuth2 or JWT authentication for existing services
- Account creation with arbitrary ACL group list

## Build

Use the Makefile in the root directory of the repository to build security-proxy-setup:

```sh
$ make cmd/security-proxy-setup/security-proxy-setup
```

This will create an executable located at `cmd/security-proxy-setup/` if successful.

## Run security-proxy-setup with different parameters

The binary supports multiple command line parameters 

```sh
 --init  // run initialization procedure for security service
 --insecureSkipVerify // skip server side SSL verification, primarily for self-signed certs
 --configfile // use different configuration file than the default
 --reset // reset proxy by removing all customizations
 --useradd // user to be added to consume the edgex services, requires 'group' parameter
 --group // group that the user belongs to
 --userdel // user to be deleted from the the proxy services
```

An example of use of the parameters can be found in the docker compose file

https://github.com/edgexfoundry/developer-scripts/blob/master/releases/fuji/compose-files/docker-compose-fuji.yml

## Docker Build

Go to the root directory of the repository and use the Makefile to build the docker container image for `security-proxy-setup`:

```sh
$ make docker_security_proxy_setup
```

It should create a docker image with the name `edgexfoundry/docker_security_proxy_setup:<version>-dev` if sucessfully built.
