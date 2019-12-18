# Configuring Microservices to use Redis

As of EdgeX 1.0 (Edinburgh), all the microservices with the exception of logging are enabled for
Redis. At this time there is no plan to add support for logging. When deploying EdgeX via Snap, the default behavior may still be Mongo (its moving to Redis). Please refer to [EdgeX snap](https://github.com/edgexfoundry/edgex-go/blob/master/snap/README.md) for more information.

When deploying EdgeX via Docker (or native), the default behavior is to use Mongo and you'll need to explicitly configure Redis.

## Requirements

### Redis

Redis can be run locally or as a Docker container. This document assumes the default port of 6379 is being used. If you are using a different port, massage these instructions appropriately.

### EdgeX

Follow the EdgeX build instructions or the EdgeX Docker Compose instructions to set yourself up for success.

## Using Docker

When starting EdgeX containers, Redis must be explicitly stated as the desired database for the microservices.

```sh
make EDGEX_DB=redis run_docker
```

This will start Redis and all the microservices.

## Using native microservices

The `configuration.toml` files for Core Data, Core Metadata, Support Notifications, and Support Scheduler found in their respective `cmd/<SERVICE>/res` directory.

For each of the microservices update the keys in the `Databases.Primary` table

| Key  | Value   |
| ---- | ------- |
| Port | 6379    |
| Type | redisdb |

Redis does not use the other keys in that table