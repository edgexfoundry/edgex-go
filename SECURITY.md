# Security Components in EdgeX Foundry Services

Starting with the Fuji release, EdgeX includes enhanced security features that are enabled by default.
There are 3 major components that are responsible for security features:

* [Security-secrets-setup](cmd/security-secrets-setup/README.md)
* [Security-secretstore-setup](cmd/security-secretstore-setup/README.md)
* [Security-proxy-setup](cmd/security-proxy-setup/README.md)

Security-secrets-setup is responsible for creating necessary certificates.
Security-secretstore-setup is responsible for initializing the secret store to hold various credentials
for EdgeX. Security-proxy-setup is responsible for initializating the EdgeX proxy environment, which
includes setting up related permissions and authentication methods. The proxy will protect all the REST
API resources.

## Get Started

To get started, fetch the latest docker-compose.yml and start the EdgeX containers:

```sh
$ wget https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/releases/nightly-build/compose-files/docker-compose-nexus.yml
$ docker-compose up -d
```

Once EdgeX is up and running, the following steps are required to access EdgeX resources:

1. The user needs to create an access token and associate every REST request with the security token
while sending the request. Use "admin" as group name below, as it is the privileged group in the default configuration of the proxy.
`<account>` below should be substituted for the desired account name (e.g., "mary", "iot_user", etc).

    ```sh
    $ docker-compose run -f docker-compose-nexus.yml --rm --entrypoint /edgex/security-proxy-setup edgexproxy --init=false --useradd=<account> --group=<groupname>
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
    curl -k https://{proxy-service-ip}:8443/coredata/api/v1/ping -H "Authorization: Bearer <access-token>"
    ```

   Note the request is made over https to the proxy-service's IP address on port 8443.  The access token is also
   included in the Authorization header.

## Community

* Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/enQtNDgyODM5ODUyODY0LWVhY2VmOTcyOWY2NjZhOWJjOGI1YzQ2NzYzZmIxYzAzN2IzYzY0NTVmMWZhZjNkMjVmODNiZGZmYTkzZDE3MTA)
* Mailing lists: [https://lists.edgexfoundry.org/mailman/listinfo](https://lists.edgexfoundry.org/mailman/listinfo)

## License

[Apache-2.0](LICENSE)
