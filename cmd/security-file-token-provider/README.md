# Integration testing instructions for security-file-token-provider

## Build

Use Makefile in the base directory of this repository to build the docker image of security-secretes-setup:

```sh
make -C ../.. cmd/security-file-token-provider/security-file-token-provider
```

It should create an executable named `security-file-token-provider` in the current directory.


## Installation

It is intended that this utility be invoked as the `tokenprovider` of `security-secretstore-setup`
after unsealing of the secret store has been completed.
