# EdgeX Security

For the most up to date information related to the EdgeX security disclosure, vulnerabilities, and hardening guide refer to the [EdgeX Security Wiki Page](https://wiki.edgexfoundry.org/display/FA/Security)

Starting with the Fuji release, EdgeX includes enhanced security features that are enabled by default.
There are 2 major components that are responsible for security features:

| Component  | Description  |
|---|---|
|  [Security-secretstore-setup](cmd/security-secretstore-setup/README.md) | Security-secretstore-setup is responsible for initializing the secret store to hold various credentials for EdgeX.  |
| [Security-proxy-setup](cmd/security-proxy-setup/README.md)  | Security-proxy-setup is responsible for initializating the EdgeX proxy environment, which includes setting up related permissions and authentication methods. The proxy will protect all the REST API resources.  |

When starting a secure EdgeX deployment, the sequence is defined by the
[EdgeX secure bootstrapping Architecture Decision Record](https://docs.edgexfoundry.org/2.0/design/adr/security/0009-Secure-Bootstrapping/).  In general the sequence is:

1. Start a bootstrapping component that sequences the framework startup.
1. Start the EdgeX secret store component.
1. Start the `secretstore-setup` container to initialize the secret store and create shared secrets (such as database passwords) needed by the microservices.
1. Start the other stateful components of EdgeX.
1. Start the non-stateful EdgeX services.
1. Start the `proxy-setup` container to configure the [Kong](https://konghq.com/) API gateway.

The startup sequence is automatically coordinated to boot in the proper order to ensure correct functioning of EdgeX.

## Get Started

This documentation assumes that EdgeX has been started from the
`docker-compose` scripts in the
[edgex-compose respository](https://github.com/edgexfoundry/edgex-compose).

Once EdgeX is up and running, the following steps are required to access EdgeX resources:

1. The user needs to create an access token and associate every REST request with the security token
   while sending the request.  Run the following command from the `edgex-compose` root folder:

    ```console
    $ make get-token
    ```

    The above command will create an access token. One example of an access token is:
    `eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjMwODMwNzgsImlhdCI6MTYyMzA3OTQ3OCwiaXNzIjoiOGJlYjRiNzUtZDA0Mi00YmE0LWFlOTctOTFjMDcxYTJmZGM0IiwibmJmIjoxNjIzMDc5NDc4fQ.KqAgCg63zjtFSvMtChLoLyIrmh8xQdb0t4sroIbLhOtgBnacaTlOdoT33VQY0QGkCEFdE1VT8WjjwrbIwitpDQ`.  
    Yours will differ from this one.

    Note that the `get-token` Makefile target by default generates a public/private keypair
    in a temporary folder that is **deleted** after the token is created.
    *For production usage, one would want to keep this keypair around
    and use it to generate future access tokens.*

2. Individual microservice ports (such as 59880, 59881 etc.) will be inaccessible for security reasons.
Instead, all the REST requests need to go through the proxy, and the proxy will redirect the request to individual microservice on behalf of the user.

    E.g, if we need to send a request to the ping endpoint of coredata, without security this would look like:

    ```sh
    curl http://{coredata-microservice-ip}:59880/api/v1/ping
    ```

    With security services enabled, the request would look like this:

    ```sh
    token=(paste output of `make get-token` here)
    curl -k https://{kong-ip}:8443/core-data/api/v2/ping -H "Authorization: Bearer $token"
    ```

   Note the request is made over https to the proxy-service's IP address on port 8443.  The access token is also
   included in the `Authorization` header.

## Starting a Non-Docker Microservice With Security Enabled

As an example, let's say you want to start core-metadata outside a container so you can debug it. 
All EdgeX microservices need a non-expired authentication token to the EdgeX secret store in order to run.
Be aware that

* Tokens expire after 60 minutes if not renewed. If you are starting/stopping a microservice and the service token has expired, fully shut down and restart the environment (`make down` and `make run`) to get fresh tokens for all.
* `/tmp/edgex/secrets/...` where the tokens live is only root readable. You can run the microservice as root or use the following to open up the tree. Note you will need to repeat the chmod each time you restart `secretstore-setup` service.

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
EDGEX_SECURITY_SECRET_STORE=true ./core-metadata
```

| Environment Override  | Description  |
|---|---|
| EDGEX_SECURITY_SECRET_STORE=true | Enable security flag (optional, security enabled by default) |

Security is enabled by default, but some developers run EdgeX =
with `EDGEX_SECURITY_SECRET_STORE=false` set in their environments.
The above command ensures that security is enabled for this particular run of the EdgeX microservice.

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

* Chat: <https://edgexfoundry.slack.com/>
* Mailing lists: [https://lists.edgexfoundry.org/mailman/listinfo](https://lists.edgexfoundry.org/mailman/listinfo)

## License

[Apache-2.0](LICENSE)
