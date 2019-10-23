# Security Components in EdgeX Foundry Services

Starting from Fuji release, EdgeX includes enhanced security features. There are 3 major components that are responsible for security
features: 

- [Security-secrets-setup] (https://github.com/edgexfoundry/edgex-go/blob/master/cmd/security-secrets-setup/README.md)
- [Security-secretstore-setup] (https://github.com/edgexfoundry/edgex-go/blob/master/cmd/security-secretstore-setup/README.md)
- [Security-proxy-setup] (https://github.com/edgexfoundry/edgex-go/blob/master/cmd/security-proxy-setup/README.md)

Security-secrets-setup is responsible for creating necessary certificates. Security-secretstore-setup is responsible for initializing the secret store that holds various credentials for EdgeX. Security-proxy-setup is responsible for initializating the environment for proxy for EdgeX to protect all the REST API resources, set up related permissions and authentication methods. 

# Get Started

To get started, fetch the latest docker-compose.yml and start the EdgeX containers:

```sh
wget https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/releases/fuji/compose-files/docker-compose-fuji-1.1.0.yml
docker-compose up -d
```

Once the EdgeX is up and running, the following steps are required to access the resources of EdgeX.

1. The user needs to create an access token and associate every REST request with the security token
while sending the request. Use "admin" as groupname below as it is the default enable group in the configuration of proxy.

```sh
docker-compose run security-proxy-setup --useradd=<account> --group=<groupname>
```

which will create an access token. The token will be like this: eyJpc3MiOiI5M3V3cmZBc0xzS2Qwd1JnckVFdlRzQloxSmtYOTRRciIsImFjY291bnQiOiJhZG1pbmlzdHJhdG9yIn0

2. The exported external ports (such as 48080, 48081 etc.) will be inaccessible due to security enhancement. Instead all the REST requests need to go through the proxy, and the proxy will redirect the request to individual microservice on behalf of the user.
E.g, if we need to send a request to the ping endpoint of coredata, the original request be like this below:

```
curl http://{coredata-microservice-ip}:48080/api/v1/ping
```

Such request needs to be converted into a request like this below:

```
curl -k https://{proxy-service-ip}:8443/coredata/api/v1/ping -H "Authorization: Bearer <access-token>"
```

# Community

- Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/enQtNDgyODM5ODUyODY0LWVhY2VmOTcyOWY2NjZhOWJjOGI1YzQ2NzYzZmIxYzAzN2IzYzY0NTVmMWZhZjNkMjVmODNiZGZmYTkzZDE3MTA)
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo

# License

[Apache-2.0](LICENSE)
