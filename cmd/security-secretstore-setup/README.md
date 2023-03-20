# EdgeX Foundry Security Service - Security Secretstore Setup

[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security-secretstore-setup service (aka edgex-vault-worker). Prior to the Ireland release, the container relies on the `security-secrets-setup` container to create the PKI, in which the requirements of TLS in a single box are no more. `security-secretstore-setup` service also fork/execs security-file-token-provider to create the tokens, and adds shared secrets to Vault itself.

## Build

Use the Makefile in the root directory of the repository to build  security-secretstore-setup:

```sh
make cmd/security-secretstore-setup/security-secretstore-setup
```

This will create an executable located at `cmd/security-secretstore-setup/` if successful.

## Run security-secretstore-setup with different parameters

The binary supports multiple command line parameters

| Parameter                         | Description                                                                                                    |
|-----------------------------------|----------------------------------------------------------------------------------------------------------------|
| -p, --profile `name`              | Indicate configuration profile other than default                                                              |
| -r, --registry                    | Indicates service should use Registry                                                                          |
| --insecureSkipVerify=`true/false` | Indicates if skipping the server side SSL cert verifcation, similar to -k of curl                              |
| --configfile=`file.yaml`          | Use a different config file (default: res/configuration.yaml)                                                  |
| --vaultInterval=`seconds`         | **Required** Indicates how long the program will pause between vault initialization attempts until it succeeds |

An example of using the parameters can be found in the following docker compose
file:
[https://github.com/edgexfoundry/developer-scripts/blob/master/releases/fuji/compose-files/docker-compose-fuji.yml](https://github.com/edgexfoundry/developer-scripts/blob/master/releases/fuji/compose-files/docker-compose-fuji.yml)

## Docker Build

Go to the root directory of the repository and use the Makefile to build the docker container image for `security-secretstore-setup`:

```sh
make docker_security_secretstore_setup
```

It should create a docker image with the name `edgexfoundry/docker_security_secretstore_setup:<version>-dev` if sucessfully built.

## Debugging Tips

* The _RevokeRootTokens_ in [`cmd/security-secretstore-setup/res/configuration.yaml`](res/configuration.yaml) controls whether the root token used to populate Vault is deleted at when edgex-vault-worker is done. If you want to debug `security-secretstore-setup`, set this to _false_:

    ```yaml
    SecretStore
    ...
      RevokeRootTokens = false
    ```

* The edgex-vault-worker uses _compose-files_vault-config_ volume to store its token. To copy the root token from edgex-vault-worker, use

    ```sh
    docker run --rm -v compose-files_vault-config:/vault/config alpine:latest cat /vault/config/assets/resp-init.json > resp-init.json
    ```

* To verify the root token

    ```sh
    docker exec -ti edgex-vault sh -l
    export VAULT_SKIP_VERIFY=true
    export VAULT_TOKEN=s.xxxxxxxxxxxxxxxx
    vault token lookup
    ```

    where `s.xxxxxxxxxxxxxxxx` is the _root_token_ member of `resp-init.json`

    Note if you are examining the vault with a non-root token (e.g. a microservice token) you must use the exact path to the key; you cannot drill down as you can with the root token.

* To explore the vault

    ```sh
    docker exec -ti edgex-vault sh -l
    export VAULT_SKIP_VERIFY=true
    export VAULT_TOKEN=s.xxxxxxxxxxxxxxxx
    vault kv list secret/
    ```

    and drill down from there. To read a key use `vault kv get` or `vault read`.

    ```sh
    docker exec -ti edgex-vault sh -l
    export VAULT_SKIP_VERIFY=true
    export VAULT_TOKEN=s.xxxxxxxxxxxxxxxx
    vault kv get /secret/edgex/redis/redis5
    ```

    Note you can set the environment variables on the docker command line with `-e` and avoid the additional shell commands.

    ```sh
    docker exec -e VAULT_SKIP_VERIFY=true ...
    ```
