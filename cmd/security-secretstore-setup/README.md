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
  -p, --profile <name>                Indicate configuration profile other than default
  -r, --registry                      Indicates service should use Registry
  --insecureSkipVerify=true/false     Indicates if skipping the server side SSL cert verifcation, similar to -k of curl
  --configfile=<file.toml>            Use a different config file (default: res/configuration.toml)
  --vaultInterval=<seconds>           Indicates how long the program will pause between vault initialization attempts until it succeeds
```

An example of using the parameters can be found in the following docker compose
file:
[https://github.com/edgexfoundry/developer-scripts/blob/master/releases/fuji/compose-files/docker-compose-fuji.yml](https://github.com/edgexfoundry/developer-scripts/blob/master/releases/fuji/compose-files/docker-compose-fuji.yml)

## Docker Build

Go to the root directory of the repository and use the Makefile to build the docker container image for `security-secretstore-setup`:

```sh
make docker_security_secretstore_setup
```

It should create a docker image with the name `edgexfoundry/docker_security_secretstore_setup:<version>-dev` if sucessfully built.
