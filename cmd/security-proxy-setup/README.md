# EdgeX Foundry Security Service - Security Proxy Setup
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

This folder builds a container that configures the NGINX reverse proxy and contains a copy of the `secrets-config` utility.

The `security-proxy-setup` binary that configured the Kong reverse proxy is removed with EdgeX 3.0

## Docker Build

Go to the root directory of the repository and use the Makefile to build the docker container image for `security-proxy-setup`:

```sh
$ make docker_security_proxy_setup
```

It should create a docker image with the name `edgexfoundry/docker_security_proxy_setup:<version>-dev` if successfully built.
