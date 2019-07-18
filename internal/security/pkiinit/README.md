# EdgeX Foundry Security Public Key Infrastructure (PKI) Init Service

[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

## Introduction

This is an implemention of the module PKI-init to initialize the PKI related materials like CA certificates, TLS certificates, and private keys for the secure secret store to protect keys, certificates and other sensitive assets for the EdgeX Foundry project. Please refer to [Security Secret Store Chapter](https://docs.edgexfoundry.org/Ch-SecretStore.html) for a detailed documentation.

The new pki-init service supports three modes of operation: In the first mode, "generate", the PKI is generated afresh every time the framework is started.  The second mode, "cache", supports a cached PKI mode where the PKI is generated once and saved to a docker volume--this saves ~1 second of startup time but leaves TLS private keys stored persistently on disk unencrypted.  An optional "cacheca" mode also caches the CA private key as well to support the possibility of after-the-fact creation of new TLS end-entity certificates.  The third mode, "import", suppresses all PKI generation functionality and tells pki-init to assume that $PKI_CACHE contains a pre-populated PKI such as a Kong certificate signed by an external certificate authority or TLS keys signed by an offline enterprise certificate authority.

## Build

For running in Docker, please build the binaries and docker images before run `docker-compose up` on the existing file `docker-compose.yml`.  To build it, run the followings:

1. In the base directory, run `make build`
2. In the base directory, run `make docker`

## Run Docker

On the command line console, run `docker-compose up --build` from the directory `pkiinit` to start the whole Docker container stack.

## Command Line Options

> * --generate: to genearte new TLS certificates
> * --import: to deploy TLS cetificates from cached area
> * --cache: generate fresh TLS certificates with erase of CA keys and put into cached area
> * --cacheca: generate fresh TLS certificates with no erase of CA keys and put into cached area
> * --help or -h: help message for the pki-init options

Please also see the [pki-init man page](pki-init.1.md).
