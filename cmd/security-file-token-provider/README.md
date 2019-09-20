# Integration testing instructions for security-file-token-provider

## Build

Use Makefile in the base directory of this repository to build the docker image of security-secretes-setup:

```sh
make -C ../.. cmd/security-file-token-provider/security-file-token-provider
```

It should create an executable named `security-file-token-provider` in the current directory.

## Run

Run the `integrationtest.sh` script.

```sh
./integrationtest.sh
```

Verify the resulting output for correctness.
The script will automatically fail if a subcommand
returns a nonzero exit status.

Afterwards, shut down the docker composition

```sh
docker-compose down
```

## Details

The integration test script first starts Vault and Consul via a docker-compose.yml script.
The script copies out the CA certificate and root token JSON file to the current directory.
It then overwrites `res/configuration.toml` with some values to allow the integration test to run.
The script then dumps a list of all policies, and a list of current Vault tokens.

Next, the script runs `security-file-token-provider`

Afterwards, the script then dumps the policies, Vault tokens,
the policy it just created, and calls the token lookup command
on the newly-created Vault token.

## Testing with TLS

In order to test with TLS, it is necessary to edit `/etc/hosts` and add
an alias for edgex-vault:

```
127.0.0.1 edgex-vault
```

Then edit `integrationtest.sh` and uncomment `CaFilePath`.

Re-run `integrationtest.sh` and verify that it makes a TLS connection to Vault.

_BE SURE TO REMOVE THE EDITS FROM `/etc/hosts` WHEN YOU ARE DONE!_
