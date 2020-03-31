# Security Components in EdgeX Foundry Services

Starting with the Fuji release, EdgeX includes enhanced security features that are enabled by default.
There are 3 major components that are responsible for security features:

| Component  | Description  |
|---|---|
| [Security-secrets-setup](cmd/security-secrets-setup/README.md)  | Security-secrets-setup is responsible for creating necessary certificates.  | 
|  [Security-secretstore-setup](cmd/security-secretstore-setup/README.md) | Security-secretstore-setup is responsible for initializing the secret store to hold various credentials for EdgeX.  |
| [Security-proxy-setup](cmd/security-proxy-setup/README.md)  | Security-proxy-setup is responsible for initializating the EdgeX proxy environment, which includes setting up related permissions and authentication methods. The proxy will protect all the REST API resources.  |

When starting a secure EdgeX deployment, the sequence is [see docker-compose-nexus-redis.yml for reference](https://github.com/edgexfoundry/developer-scripts/blob/master/releases/nightly-build/compose-files/docker-compose-nexus-redis.yml))

1. Start the `edgex-secrets-setup` container from the `docker-edgex-secrets-setup-go` image to create the PKI.
1. Start [Vault by HashiCorp](https://www.vaultproject.io/)
1. Start the `edgex-vault-worker` container from the `docker-edgex-security-secretstore-setup-go` image to create the shared secrets needed by the microservices.
1. Finally, the start the `edgex-proxy` container from the `docker-edgex-security-proxy-setup-go` image once [Kong](https://konghq.com/) is up.

## Get Started

To get started, fetch the latest [docker-compose-nexus-redis.yml](https://github.com/edgexfoundry/developer-scripts/blob/master/releases/nightly-build/compose-files/docker-compose-nexus-redis.yml)) and start the EdgeX containers:

```sh
wget https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/releases/nightly-build/compose-files/docker-compose-nexus-redis.yml
docker-compose up -d
```

Once EdgeX is up and running, the following steps are required to access EdgeX resources:

1. The user needs to create an access token and associate every REST request with the security token
   while sending the request. Use _admin_ as group name, as it is the privileged group in the
   default configuration of the proxy. Use anything for _user_ as the desired account name
   (e.g., "mary", "iot_user", etc).

    ```sh
    docker-compose -f docker-compose-nexus-redis.yml run --rm --entrypoint /edgex/security-proxy-setup edgex-proxy --init=false --useradd=IAmGroot --group=admin
    ```

    which will create an access token. One example of an access token is:
    `eyJpc3MiOiI5M3V3cmZBc0xzS2Qwd1JnckVFdlRzQloxSmtYOTRRciIsImFjY291bnQiOiJhZG1pbmlzdHJhdG9yIn0`.  
    Yours will differ from this one.

2. The exported external ports (such as 48080, 48081 etc.) will be inaccessible for security reasons.
Instead, all the REST requests need to go through the proxy, and the proxy will redirect the request to individual microservice on behalf of the user.

    E.g, if we need to send a request to the ping endpoint of coredata, without security this would look like:

    ```sh
    curl http://{coredata-microservice-ip}:48080/api/v1/ping
    ```

    With security services enabled, the request would look like this:

    ```sh
    curl -k https://{kong-ip}:8443/coredata/api/v1/ping -H "Authorization: Bearer <access-token>"
    ```

   Note the request is made over https to the proxy-service's IP address on port 8443.  The access token is also
   included in the Authorization header.

## Starting a Non-Docker Microservice With Security Enabled

As an example, let's say you want to start core-metadata outside a container so you can debug it. You will need a non-expired token to authenticate and authorize the service and you will need to tell core-metadata about its environment. Be aware that

* Tokens expire after 60 minutes if not renewed. If you are starting/stopping a microservice and the service token has expired, stop and start security-secretstore-setup (aka vault-worker).
* `/tmp/edgex/secrets/...` where the tokens live is only root readable. You can run the microservice as root or use the following to open up the tree. Note you will need to repeat the chmod each time you restart `security-secrets-setup`

    ```sh
    pushd /tmp/edgex
    sudo find . -type d -exec chmod a+rx {} /dev/null \;
    sudo find . -type f -exec chmod a+r {} /dev/null \;
    popd
    ```

Fortunately, between go-mod-boostrap and [microservice self seeding](https://github.com/edgexfoundry/edgex-docs/blob/master/docs_src/design/adr/0005-Service-Self-Config.md) this is quite straight forward:

```sh
sudo bash

cd cmd/core-metadata
SecretStore_ServerName=edgex-vault SecretStore_RootCaCertPath=/tmp/edgex/secrets/ca/ca.pem SecretStore_TokenFile=/tmp/edgex/secrets/edgex-core-metadata/secrets-token.json Logging_EnableRemote="false" ./core-metadata
```

| Environment Override  | Description  |
|---|---|
| SecretStore_ServerName=edgex-vault | Name of secret store to compare against X.509 DN |
| SecretStore_RootCaCertPath=/tmp/edgex/secrets/ca/ca.pem | Root CA for validating TLS certificate chain |
| SecretStore_TokenFile=/tmp/edgex/secrets/edgex-core-metadata/secrets-token.json | Authorization token |
| Logging_EnableRemote="false" | There can only one (logger) |

The defaults are loaded from `res/configuration.toml` and then the environment variables override the defaults.

* Both Kong and the microservice need to agree on the host were the microservice lives. To that end, if you are mixing container based infrastructure (e.g. Kong) and native microservices you will need to do the following
  * Tell the Kong container about the host (assume its 10.138.0.2 in this example). Add the following to the Kong entry in the compose file
  
  ```yaml
  extra_hosts:
      - "edgex-core-metadata:10.138.0.2"
      - "edgex-core-data:10.138.0.2"
      - "edgex-core-command:10.138.0.2"
  ```

  and add the following to the environment section

  ```yaml
  - 'Clients_CoreData_Host: edgex-core-data'
  - 'Clients_Logging_Host: edgex-support-logging'
  - 'Clients_Notifications_Host: edgex-support-notifications'
  - 'Clients_Metadata_Host: edgex-core-metadata'
  - 'Clients_Command_Host: edgex-core-command'
  - 'Clients_Scheduler_Host: edgex-support-scheduler'
  ```

  * Tell the OS about the host by adding the following to `/etc/hosts`

  ```sh
  10.138.0.2    edgex-core-data
  10.138.0.2    edgex-core-metadata
  10.138.0.2    edgex-core-command
  ```

  and tell the microservice about the host (only core-metadata shown here)

  ```sh
  export Clients_Metadata_Host=edgex-core-metadata
  ```

## Community

* Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/enQtNDgyODM5ODUyODY0LWVhY2VmOTcyOWY2NjZhOWJjOGI1YzQ2NzYzZmIxYzAzN2IzYzY0NTVmMWZhZjNkMjVmODNiZGZmYTkzZDE3MTA)
* Mailing lists: [https://lists.edgexfoundry.org/mailman/listinfo](https://lists.edgexfoundry.org/mailman/listinfo)

## License

[Apache-2.0](LICENSE)
