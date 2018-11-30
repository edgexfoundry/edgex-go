# Configuring Microservices to use Redis

As of Delhi 0.7.1, Core Data, Core Metadata, and Export Client all support using Redis as their persistence layer. The microservices can be configured via Consul or in a development environment, via the respective TOML files.

## Requirements

### Redis

Redis can be run locally or as a Docker container. This document assumes the default port of 6379 is being used. If you are using a different port, massage these instructions appropriately.

### EdgeX

Follow the EdgeX build instructions or the EdgeX Docker Compose instructions to set yourself up for success.

## Configure via Consul

### Start EdgeX

```sh
export EDGEX_CORE_DB=redis
make run_docker
```

This will start all the microservices along with both Redis and Mongo

### Configure Microservices to use Redis

1. Point your browser at Consul. For example `http://localhost:8500/ui`
2. Modify the primary database service key for each of edgex-core-data, edgex-core-metadata, and edgex-export-client

- edgex-core-data

<table align="center">
    <tr>
        <td align="left">Consul Path</td>
        <td align="left">edgex-core-data</td>
    </tr>
    <tr>
        <td align="left">Host</td>
        <td align="left">edgex-redis</td>
    </tr>
    <tr>
        <td align="left">Port</td>
        <td align="left">6379</td>
    </tr>
    <tr>
        <td align="left">Type</td>
        <td align="left">redisdb</td>
    </tr>
</table>

- edgex-core-metadata

  <table align="center">
      <tr>
          <td align="left">Consul Path</td>
          <td align="left">edgex-core-meadata</td>
      </tr>
      <tr>
          <td align="left">Host</td>
          <td align="left">edgex-redis</td>
      </tr>
      <tr>
          <td align="left">Port</td>
          <td align="left">6379</td>
      </tr>
      <tr>
          <td align="left">Type</td>
          <td align="left">redisdb</td>
      </tr>
  </table>

- edgex-export-client
  <table align="center">
      <tr>
          <td align="left">Consul Path</td>
          <td align="left">edgex-core-meadata</td>
      </tr>
      <tr>
          <td align="left">Host</td>
          <td align="left">edgex-redis</td>
      </tr>
      <tr>
          <td align="left">Port</td>
          <td align="left">6379</td>
      </tr>
      <tr>
          <td align="left">Type</td>
          <td align="left">redisdb</td>
      </tr>
  </table>

3. Restart the Microservices

```sh
docker restart edgex-core-data
docker restart edgex-core-metadata
docker restart edgex-export-client
```

## Configure via TOML files

The `configuration.toml` files for edgex-core-data, edgex-core-metadata, and edgex-export-client are found in their respective `cmd/<SERVICE>/res` directory

For each of the microservices update the keys in the `Databases.Primary` table

| Key  | Value   |
| ---- | ------- |
| Port | 6379    |
| Type | redisdb |

Redis does not use the other keys in that table