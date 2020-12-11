# EdgeX Foundry Security Service - Security Bootstrap Redis

[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

## Summary

Implements a service to read the Redis password from Vault and set it in Redis. The service exits when it completes. The Docker entrypoint keeps the image from exiting until it receives an interrupt.

## Explanation

In the Geneva release, this code was liberally borrowed from [edgexfoundry/docker-edgex-mongo](https://github.com/edgexfoundry/docker-edgex-mongo), tightly coupled with security-secretstore-setup, and used a shared file to pass the Redis password to Redis. For the Hanoi release, the goal is to conform to the [ADR for secret creation and distribution](https://docs.edgexfoundry.org/1.2/design/adr/security/0008-Secret-Creation-and-Distribution/) and support configuration overrides via go-mod-boostrap.

The service is organized in the Docker compose file to run after security-secretstore-setup and Redis are started. This isn't guaranteed by Docker so the security-bootstrap-redis will keep retrying until the retry timer expires. The service starts by reading the Redis password from the vault, creates a connection to Redis (which when it starts does not require authentication), and attempts to set the password obtained from the vault.

If security-bootstrap-redis cannot create an unauthenticated connection to Redis, it will attempt to create an authenticated connection using the credentials received from vault. It is an error if this authenticated connection cannot be established as it means Redis is out of sync with the vault.

The service does not exit when started via the Docker.

## Tight Coupling

* res/configuration.toml and redis/config/config.go
* res-file-token-provider/configuration.toml and clients.SecurityBootstrapRedisKey ("edgex-security-bootstrap-redis")
* security-secretstore-setup and vault key layout
