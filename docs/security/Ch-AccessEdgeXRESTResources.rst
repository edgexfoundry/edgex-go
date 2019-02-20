############################
Access EdgeX REST resources 
############################

When the EdgeX API Gateway is used, access to the micro service APIs must go through the reverse proxy.  Requestors of an EdgeX REST endpoints must therefore change the URL they use to access the services.  Example below explain how to map the non-secured micro service URLs with reverse proxy protected URLS.
To access the ping endpoint of an EdgeX micro service (using the command service as an example), the URL is http://edgex-command-service:48082/api/v1/ping
With API gateway serving as the single access point for the EdgeX services, the ping URL is https://api-gateway-server:8443/command/api/v1/ping?jwt=<JWT-Token>
Please notice that there are 4 major differences when comparing the URLs above
1.	Switch from http to https as the API Gateway server enables https 
2.	The host address and port are switched from original micro service host address and port to a common api gateway service address and 8443 port as the api gateway server will serve as the single point for all the EdgeX services
3.	Use the name of the service (in this case “command”) within the URL to indicate that the request is to be routed to the appropriate EdgeX service (command in this example)
4.	Add a JWT as part of the URL as all the REST resources are protected by either OAuth2 or JWT authentication. The JWT can be obtained when a user account is created with the security API Gateway. 
