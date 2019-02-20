###################
API Gateway
###################

The security API gateway is the single point of entry for all EdgeX REST traffic. It is the barrier between external clients and the EdgeX microservices preventing unauthorized access to EdgeX REST APIs. The API gateway accepts client requests, verifies the identity of the clients, redirects the requests to correspondent microservice and relays the results back to the client. 

The API Gateway provides an HTTP REST interface for administration management. The administrative management offers the means to configure API routing, as well as client authentication and access control.  This configuration is store in an embedded database. 

KONG (https://konghq.com/) is the product underlying the API gateway.  The EdgeX community has added code to initialize the KONG environment, set up service routes for EdgeX microservices, and add various authentication/authorization mechanisms including JWT authentication, OAuth2 authentication and ACL.


======================
Start the API Gateway
======================

Start the API gateway with Docker Compose and a Docker Compose manifest file (the Docker Compose file named docker-compose-delhi-0.7.0.yml found at https://github.com/edgexfoundry/developer-scripts/tree/master/compose-files/security)).  This Compose file starts all of EdgeX including the security services. The command to start EdgeX inclusive of API gateway related services is:
::

    docker-compose up -d

For debugging purpose, the API gateway services can be started individually with these commands used in sequence after secret store starts successfully. Lines starts with # are comments to explain the purpose of the command.
::

    docker-compose up -d kong-db
    # start up backend database for API gateway

    docker-compose up -d kong-migrations
    # initialize the backend database for API gateway

    docker-compose up -d kong
    # start up KONG the major component of API gateway

    docker-compose up -d edgex-proxy
    # initialize KONG, configure proxy routes, apply certificates to routes, and enable various authentication/ACL features. 


If the last command returns an error message for any reason (such as incorrect configuration file), the API gateway may be in an unstable status. The following command can be used to stop and remove the containers. 
::

    docker-compose down
    # stop and remove the containers

After stopping and removing the containers, you can attempt to recreate and start them again. Alternatively you can use the command to reset the API gateway as shown below:
::

    docker run –network=edgex-network edgexfoundry/docker-edgex-proxy-go --reset=true 

After issuing the reset command, attempt to start and reinitialize with the command below.
::

    docker run –network=edgex-network edgexfoundry/docker-edgex-proxy-go --init=true 

You can learn more about these commands, to include some additional options by running: 
::

    docker run –network=edgex-network edgexfoundry/docker-edgex-proxy-go –h 

On successfully starting EdgeX with the API Gateway services, the list of running containers should close follow the listing shown below.  Note key security service containers like kong, kong-db, edgex-vault are listed. 

.. image:: Running.Security.png


========================
Configuring API Gateway
========================

The API gateway supports two different forms of authentication: JSON Web Token (JWT) or OAuth2 Authentication. Only one authentication method can be enabled at a time. The API Gateway also supports an Access Control List (ACL) which can be enabled with one of the authentication methods mentioned earlier for fine control among the groups. The authentication and ACL need to be specified in the API gateway’s configuration file.  Setup of authentication and access control occurs automatically as part of API gateway initialization. The configuration file can be found at https://github.com/edgexfoundry/security-api-gateway/blob/master/core/res/configuration-docker.toml


**Configuration of JWT Authentication for API Gateway**

When using JWT Authentication, the [kongauth] section needs to be specified in the configuration file as shown below.
::

    [kongauth]
    name = "jwt"


**Configuration of OAuth2 Authentication for API Gateway**

When using OAuth2 Authentication, the [kongauth] section needs to specify oauth2 in the configuration file as shown below. Note, today EdgeX only supports “client credential” authentication (specified in “granttype”) currently for OAuth. 

::

    [kongauth]
    name = "oauth2"
    scopes = "email,phone,address"
    mandatoryscope = "true"
    enableclientcredentials = "true"
    clientid = "test"
    clientsecret = "test"
    redirecturi = "http://edgex.com"
    granttype = "client_credentials"
    scopegranted = "email"
    resource = "coredata"
    

**Configuration of ACL for API Gateway**

Access control is also specified in the configuration file as shown below.  Note, users that belong to the whitelist will have access to the resources of EdgeX, and users not belonging to the group listed here will be denied when trying to access resources through the API Gateway.
::

    [kongacl]
    name = "acl"
    whitelist = "admin,user"
    

======================
Using API Gateway
======================

**Resource Mapping between EdgeX Microservice and API gateway**


If the EdgeX API gateway is not in use, a client can access and use any REST API provided by the EdgeX microservices by sending an HTTP request to the service endpoint. E.g., a client can consume the ping endpoint of the Core Data microservice with curl command like this:
::

    curl http://<core-data-microservice-ip>:48080/api/v1/ping

Once the API gateway is started and initialized successfully, and all the common ports for EdgeX microservices are blocked by disabling the exposed external ports of the EdgeX microservices through updating the docker compose file, the EdgeX microservice will be behind the gateway.  At this time both the microservice host/IP Address (<core-data-microservice-ip> in the example) as well as the service port (48080 in the example) are not available to external access. EdgeX uses the gateway as a single entry point for all the REST APIs. With the API gateway in place, the  curl command to ping the endpoint of the same Core Data service, as shown above, needs to change to : 
::

    curl https://<api-gateway-host-ip>:8443/coredata/api/v1/ping

Comparing these two curl commands you may notice several differences.

*	“Http” is switched to "https" as we enable the SSL/TLS for secure communication.  This applies to any client side request.
*	The EdgeX microservice IP address where the request is sent changed to the host/IP address of API gateway service (recall the API gateway becomes the single entry point for all the EdgeX micro services). The API gateway will eventually lateral the request to the Core Data service if the client is authorized.  
*	The port of the request is switched from 48080 to 8443, which is the default SSL/TLS port for API gateway (versus the micro service port). This applies to any client side request. 
*	The "/coredata/" path in the URL is used to identify which EdgeX micro service the request is routed to.  As each EdgeX micro service has a dedicated service port open that accepts incoming requests, there is a mapping table kept by the API gateway that maps paths to micro service ports.   A partial listing of the map between ports and URL paths is shown in the table below.


````

+-------------------------+-------------+----------------+
| EdgeX microservice Name | Port number | Partial URL    |
+-------------------------+-------------+----------------+
| coredata                | 48080       | coredata       |
+-------------------------+-------------+----------------+
| metadata                | 48081       | metadata       |
+-------------------------+-------------+----------------+
| command                 | 48082       | command        |
+-------------------------+-------------+----------------+
| notifications           | 48060       | notifications  |
+-------------------------+-------------+----------------+
| supportlogging          | 48061       | supportlogging |
+-------------------------+-------------+----------------+
| exportdistro            | 48070       | exportdistro   |
+-------------------------+-------------+----------------+
| exportclient            | 48071       | exportclient   |
+-------------------------+-------------+----------------+

````

**Creating Access Token for API Gateway Authentication**


If the EdgeX API gateway is not in use, a client can access and use any REST API provided by the EdgeX microservices by sending an HTTP request to the service endpoint. E.g., a client can consume the ping endpoint of the Core Data microservice with curl command like this:
::

    curl http://<core-data-microservice-ip>:48080/api/v1/ping

Again, the request doesn’t include client identity information. Once the API gateway is started and initialized successfully, the EdgeX microservice REST APIs will be behind the gateway, an access token must  be attached with any client-side HTTP request for identity verification and authentication done by the API gateway. This access token is different from the access token of secret store even though they have the same name. The purpose of the access token for the API gateway is to identity clients that send the requests to consume the REST API of EdgeX. The secret store will then use the token to verify the identity of clients that send the request to access the secrets of EdgeX that are stored in the secret store.
To obtain an access token for a client, a user that is associated with the client as well as a group that the user belongs to needs to be added into the API gateway. The command to add a user and the group is: 
::

    docker run –network=edgex-network edgexfoundry/docker-edgex-proxy-go --useradd=<user> --group=<groupname> 

The command above will return an access token that can then be used by the client to access the EdgeX REST API resources. Depending on the choice of authentication method, the format of the access token will be something like this if JWT is enabled: 
::
        
 eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiI5M3V3cmZBc0xzS2Qwd1JnckVFdlRzQloxSmtYOTRRciIsImFjY291bnQiOiJhZG1pbmlzdHJhdG9yIn0.em8ffitqrd59_DeYKfQkTZGtUA1T99NikETwtedOgHQ

Alternatively, the access token may look like what is shown below if the OAuth2 is enabled:
::

    MNsBh6jDDSxaECzUtimW1nDSvI2v0xsZ

If a client needs to be disabled and the client’s token invalidated, use the command here to remove/delete the user: 
::

    docker run –network=edgex-network edgexfoundry/docker-edgex-proxy-go –-userdel=<user>


**Using API Gateway to Proxy Existing EdgeX Microservices**


Once the resource mapping and access token to API gateway are in place, a client can use the access token to use the protected EdgeX REST API resources behind the API gateway.
Again, without the API Gateway in place, here is the sample request to hit the ping endpoint of the EdgeX Core Data microservice using curl: 
::

    curl http://<core-data-microservice-ip>:48080/api/v1/ping

With the security service and JWT authentication is enabled, the command changes to:
::

    curl –H “host: edgex” https://<api-gateway-service-ip>:8443/coredata/v1/ping?  -H "Authorization: Bearer <access-token>”

In summary the difference between the two commands are listed below:

*	–H "host: edgex" is used to indicate that the request is for EdgeX domain as the API gateway could be used to take requests for different domains. 
*	Use the https versus http protocol identifier for SSL/TLS secure communication.
*	The service port 8443 is the default TLS service port of API gateway
*	Use the URL path "coredata" to indicate which EdgeX microservice the request is routed to
*	Use header of -H "Authorization: Bearer <access-token>" to specify the access token associated with the client that was generated when the client was added.  

The format for OAuth2 authentication is similar. For OAuth2 use the bearer token from OAuth2 authentication instead of the JWT token.  Here is an example of the curl command using OAuth2:
::

    curl –H "host: edgex" https://<api-gateway-service-ip>:8443/coredata/v1/ping -H "Authorization:bearer <access-token>"

