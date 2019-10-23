# EdgeX Foundry Security Service - Security Secretstore Setup
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security-secretstore-setup service.

## Build

Use the Makefile in the root directory of the repository to build  security-secrets-setup:

```sh
 make cmd/security-secretstore-setup/security-secretstore-setup
```

This will create an executable located at `cmd/security-secretstore-setup/` if successful.

## Run security-secretstore-setup with different parameters

The binary supports multiple command line parameters 

```sh
 --init  //run init procedure for security service
 --insecureSkipVerify //skip server side SSL verification, mainly for self-signed cert
 --configfile //use different configuration file other than the default
 --vaultInterval //time to wait between checking Vault status in seconds
```

An example of use of the parameters can be found from docker compose file

https://github.com/edgexfoundry/developer-scripts/master/releases/fuji/compose-files/docker-compose-fuji-1.1.0.yml

## Docker Build

Go to the root directory of the repository and use the Makefile to build the docker container image for `security-secretstore-setup`:

```sh
make docker_security_secretstore_setup
```

It should create a docker image with the name `edgexfoundry/docker_security_secretstore_setup:<version>-dev` if sucessfully built.
