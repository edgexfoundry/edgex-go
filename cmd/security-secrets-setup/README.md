# Testing Instruction for security-secrets-setup

## Build

Use Makefile in the base directory of this repository to build the docker image of security-secretes-setup:

```sh
make docker_security_secrets_setup
```

It should create a docker image with the name `edgexfoundry/docker-edgex-vault:<version>-dev` if sucessfully built.

# Run

To see what the TLS materials are generated inside the docker container, 
one can run the built docker image above (assuming the version number is `1.1.0`) as:

```sh
docker run --rm -it edgexfoundry/docker-edgex-vault:1.1.0-dev /bin/sh
```

once that is run, one should see the execution console outputs and gives the interactive console prompt in the base directory `/vault`.  Based on the current JSON configuration, TLS materials should be able to found in the directory of `./config/pki/EdgeXFoundryCA/` showed as follows:

```sh
/vault # ls ./config/pki/EdgeXFoundryCA/
EdgeXFoundryCA.pem    edgex-kong.pem        edgex-kong.priv.key   edgex-vault.pem       edgex-vault.priv.key
```
