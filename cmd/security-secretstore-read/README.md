# EdgeX Foundry Security Service - Security Secret Store Read
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Implements a helper program to read the Redis5 password so it can be passed to Redis when it starts. This program is used by the `security-secretstore-setup` container.

See [security-secretstore-setup/entrypoint.sh](../security-secretstore-setup/entrypoint.sh) where it is used. Please also note the dependency on the Docker named volume `db-secrets` in releases/nightly-build/compose-files/docker-compose-nexus-redis.yml from [developer-scripts](https://github.com/edgexfoundry/developer-scripts).
