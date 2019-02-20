########
Security
########

.. image:: EdgeX_Security.png

Security elements, both inside and outside of EdgeX Foundry, protect the data and control of devices, sensors, and other IoT objects managed by EdgeX Foundry. Based on the fact that EdgeX is a “vendor-neutral open source software platform at the edge of the network”, the EdgeX security features are also built on a foundation of open interfaces and pluggable, replaceable modules. 
With security service enabled, the administrator of the EdgeX would be able to initialize the security components, set up running environment for security services, manage user access control, and create JWT( JSON Web Token) for resource access for other EdgeX business services. There are two major EdgeX security components. The first is a security store, which is used to provide a safe place to keep the EdgeX secrets. The second is an API gateway, which is used as a reverse proxy to restrict access to EdgeX REST resources and perform access control related works. 
In summary, the current features are as below:

 * Secret creation, store and retrieve (password, cert, access key etc.)
 * API gateway for other existing EdgeX microservice REST APIs
 * User account creation with optional either OAuth2 or JWT authentication
 * User account with arbitrary Access Control List groups (ACL)

.. toctree::
   :maxdepth: 1

   Ch-SecretStore 
   Ch-APIGateway 



