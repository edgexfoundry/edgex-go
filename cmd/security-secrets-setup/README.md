# EdgeX Foundry Security Service - Security Secrets Setup
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security-secrets-setup service.

## Build

Use the Makefile in the root directory of the repository to build  security-secrets-setup:

```sh
make cmd/security-secrets-setup/security-secrets-setup
```

This will create an executable located at `cmd/security-secrets-setup/` if successful.

## Run **generate** subcommand

The **generate** subcommand always creates a new PKI

Before **generate** subcommand is issued, make sure *DeployDir* in the configuration toml file (in this example, it is `/run/edgex/secrets/`) exists and is writable.

Also, set *CertConfigDir* in the configuration toml to the path that holds the JSON configuration files for each certificate that will be generated.

To use the **generate** subcommand, go to `cmd/security-secrets-setup/` and then run

```sh
sudo -E ./security-secrets-setup generate
```

The `sudo` is used because the base of the deploy directory `/run/` is owned by `root:root`.

Verify that it runs successfully with generating both Vault and Kong's TLS assets under `/run/edgex/secrets/` directory.

## Run **cache** subcommand

The **cache** subcommand creates a new PKI the first time and then reuses the cached PKI on subsequent invocations.

Before **cache** subcommand is issued, make sure *CacheDir* in the configuration toml file (in this example, it is `/etc/edgex/pki`) exists and is writable.

To use the **cache** subcommand, from the same `cmd/security-secrets-setup/`  directory run

```sh
sudo -E ./security-secrets-setup cache
```

in which the configured *CacheDir* is used to specify the path to the cached location.  The `sudo` is used because the base of the deploy directory `/run/` is owned by `root:root`.

After being successfully run, the Vault and Kong TLS assets are generated and cached into the *CacheDir* `/etc/edgex/pki` and also deployed from the cache directory into the deploy directory `/run/edgex/secrets`.  One can see the files with `ls` command for example.

One can also use some file diff tool to compare the TLS files like:

```sh
sudo diff -r /etc/edgex/pki/ca/ca.pem /run/edgex/secrets/ca/ca.pem
```

After the certificates are deployed, `security-secrets-setup` will generate a  *sentinel* file `.security-secrets-setup.complete` per certificate directory in the root deploy directory to indicate successful cache.  For example:

```sh
sudo diff -r /etc/edgex/pki/ /run/edgex/secrets/

Only in /run/edgex/secrets/ca: .security-secrets-setup.complete
Only in /run/edgex/secrets/edgex-kong: .security-secrets-setup.complete
Only in /run/edgex/secrets/edgex-vault: .security-secrets-setup.complete
```

## Run **import** subcommand

To use the **import** subcommand, from the same `cmd/security-secrets-setup/` directory run

```sh
sudo -E ./security-secrets-setup import
```

The `sudo` is used because the base of the deploy directory `/run/` is owned by `root:root`.

The **import** subcommand operates differently depending on the status of cache directory:

- *CacheDir* is empty:

    In this case, import just errors out: the TLS assets are required to be in the cache directory before they can be deployed into directory `/run/edgex/secrets/`.

- *CacheDir* has already cached or been loaded with generated TLS assets before running **import** subcommand:

    One should see the deployment of TLS assets from the cache directory `/etc/edgex/pki` into the deploy directory `/run/edgex/secrets`.  One can verify that there are Vault and Kong TLS assets deployed from cache directory `/etc/edgex/pki` into the directory `/run/edgex/secrets` via `ls` command.

## Docker Build

Go to the root directory of the repository and use the Makefile to build the docker container image for `security-secrets-setup`:

```sh
make docker_security_secrets_setup
```

It should create a docker image with the name `edgexfoundry/docker-edgex-vault:<version>-dev` if sucessfully built.

## Docker Run

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

## Reference

See also [security-secrets-setup man page](https://github.com/edgexfoundry/edgex-docs/blob/master/security/security-secrets-setup.1.rst)
