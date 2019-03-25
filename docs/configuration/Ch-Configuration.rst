##########################
Configuration and Registry
##########################

============
Introduction
============

The purpose of this section is to describe the configuration and service registration capabilities of the EdgeX Foundry platform. In all cases unless otherwise specified, the examples provided are based on the reference architecture built using the `Go programming language <https://golang.org/>`_
.

-------------
Configuration
-------------

Local Configuration
"""""""""""""""""""

Because EdgeX Foundry may be deployed and run in several different ways, it is important to understand how configuration is loaded and from where it is sourced. Referring to the cmd directory within the `edgex-go repository <https://github.com/edgexfoundry/edgex-go>`_
, each service has its own folder. Inside each service folder there is a ``res`` directory (short for "resource"). There you will find the configuration files in `TOML format <https://github.com/toml-lang/toml>`_
that defines each service's configuration. A service may support several different configuration profiles, such as a "docker" profile. In this case, the configuration file located directly in the ``res`` directory should be considered the default configuration profile. Sub-directories will contain configurations appropriate to the respective profile.

With the exception of the ``config-seed`` service, which will be discussed in a moment, a service's configuration profile can be indicated using one of the following command line flags:

``--profile / -p``

Taking the ``core-data`` service as an example:

- ``./core-data``  starts the service using the default profile found locally
- ``./core-data --profile=docker``  starts the service using the docker profile found locally

Seeding Configuration
"""""""""""""""""""""

When utilizing the registry to provide centralized configuration management for the EdgeX Foundry microservices, it is necessary to seed the required configuration before starting the services. This is the responsibility of the ``config-seed`` service. The ``config-seed`` service will assume that a service registry is being used and that the necessary endpoint is included in the local configuration file. The use of profiles is supported by the ``config-seed`` service in the same manner described above.

So for example, if we wanted to seed the registry with docker-related configuration information, we could execute the following after starting the registry:

``./config-seed --profile=docker``

Assuming a successful run, the ``config-seed`` will populate the necessary values and then exit. In order for a service to now load the configuration from the registry, we must use one of the following flags:

``--registry / -r``

Again, taking the ``core-data`` service as an example:

``./core-data --registry --profile=docker`` will start the service using configuration values found in the registry

Note that when utilizing the registry, it is optional to also specify the ``--profile / -p`` flag if you are using a profile other than the default. This is because the location of the registry must still be obtained from the local config file. At this time, use of multiple profiles at once is not supported.

Configuration Structure
"""""""""""""""""""""""

Configuration information is organized into a hierarchical structure allowing for a logical grouping of services, as well as versioning, beneath an “edgex” namespace at root level of the configuration tree. The root namespace separates EdgeX Foundry-related configuration information from other applications that may be using the same registry. Below the root, sub-nodes facilitate grouping of device services, EdgeX core services, security services, etc. As an example, the top-level nodes shown when one views the configuration registry might be as follows:

- edgex *(root namespace)*
    - core *(edgex core services)*
    - devices *(device services)*
    - security *(security services)*

Versioning
""""""""""

Incorporating versioning into the configuration hierarchy looks like this.

- edgex *(root namespace)*
    - core *(edgex core services)*
        - 1.0
            - edgex-core-command
            - edgex-core-data
            - edgex-core-metadata
        - 2.0
    - devices *(device services)*
        - 1.0
            - mqtt-c
            - mqtt-go
            - modbus-go
        - 2.0

The versions shown correspond to major versions of the given services. These are not necessarily equated with long term support (LTS) releases. For all minor/patch versions associated with a major version, the respective service keys live under the major version in configuration (such as 1.0). Changes to the configuration structure that may be required during the associated minor version development cycles can only be additive. That is, key names will not be removed or changed once set in a major version, nor will sections of the configuration tree be moved from one place to another. In this way backward compatibility for the lifetime of the major version is maintained.

An advantage of grouping all minor/patch versions under a major version involves end-user configuration changes that need to be persisted during an upgrade. The ``config-seed`` will not overwrite existing keys when it runs unless explicitly told to do so. Therefore if a user leaves their configuration registry running during an EdgeX Foundry upgrade, only the new keys required to support the point release will be added to their configuration, leaving any customizations in place.

Readable vs Writable Settings
"""""""""""""""""""""""""""""

Within a given service's configuration, there are keys whose values can be edited and change the behavior of the service while it is running versus those that are effectively read-only. These writable settings are grouped under a given service key. For example, the top-level groupings for edgex-core-data are:

- /edgex/core/1.0/edgex-core-data/Clients
- /edgex/core/1.0/edgex-core-data/Databases
- /edgex/core /1.0/edgex-core-data/Logging
- /edgex/core/1.0/edgex-core-data/MessageQueue
- /edgex/core/1.0/edgex-core-data/Registry
- /edgex/core/1.0/edgex-core-data/Service
- /edgex/core/1.0/edgex-core-data/Writable

Any configuration settings found in the ``Writable`` section shown above may be changed and affect a service's behavior without a restart. Any modifications to the other settings would require a restart.

========
Registry
========

The registry refers to any platform you may use for service discovery and centralized configuration management. For the EdgeX Foundry reference implementation, the default provider for both of these responsibilities is HashiCorp's `Consul <https://www.consul.io/>`_
. Integration with the registry is handled through the `go-mod-registry <https://github.com/edgexfoundry/go-mod-registry>`_
module referenced by all services.

.. image:: EdgeX_RegistryHighlighted.png

------------------------
Introduction to Registry
------------------------

The objective of the registry is to enable microservices to find and to communicate with each other.  When each microservice starts up, it registers itself with the registry, and the registry continues checking its availability periodically via a specified health check endpoint. When one microservice needs to connect to another one, it connects to the registry to retrieve the available host name and port number of the target microservice and then invokes the target microservice. The following figure shows the basic flow.

.. image:: EdgeX_ConfigurationRegistry.png

Consul is the default registry implementation and provides native features for service registration, service discovery, and health checking.  Please refer to the Consul official web site for more information:

    https://www.consul.io

Physically, the "registry" and "configuration" management services are combined and running on the same Consul server node.

------------------
Web User Interface
------------------

A web user interface is also provided by Consul natively.  Users can view the available service list and their health status through the web user interface.  The web user interface is available at the /ui path on the same port as the HTTP API.  By default this is http://localhost:8500/ui.  For more detail, please see:

    https://www.consul.io/intro/getting-started/ui.html

-----------------
Running on Docker
-----------------

For ease of use to install and update, the microservices of EdgeX Foundry are also published as Docker images onto Docker Hub, including Registry:

    https://hub.docker.com/r/edgexfoundry/docker-core-consul/

After the Docker engine is ready, users can download the latest Consul image by the docker pull command:

    docker pull edgexfoundry/docker-core-consul

Then, startup Consul using Docker container by the Docker run command:

    docker run -p 8400:8400 -p 8500:8500 -p 8600:8600 --name edgex-core-consul --hostname edgex-core-consul -d edgexfoundry/docker-core-consul

These are the command steps to start up Consul and import the default configuration data:

1. login to Docker Hub:

  $ docker login

2. A Docker network is needed to enable one Docker container to communicate with another. This is preferred over use of --links that establishes a client-server relationship:

  $ docker network create edgex-network

3. Create a Docker volume container for EdgeX Foundry:

  $ docker run -it --name edgex-files --net=edgex-network -v /data/db -v /edgex/logs -v /consul/config -v /consul/data -d edgexfoundry/docker-edgex-volume

4. Create the Consul container:

  $ docker run -p 8400:8400 -p 8500:8500 -p 8600:8600 --name edgex-core-consul --hostname edgex-core-consul --net=edgex-network --volumes-from edgex-files -d edgexfoundry/docker-core-consul

5. Verify the result: http://localhost:8500/ui

------------------------
Running on Local Machine
------------------------

To run Consul on the local machine, requires the following steps:

1. Download the binary from Consul official website: https://www.consul.io/downloads.html.  Please choose the correct binary file according to the operation system.
2. Set up the environment variable.  Please refer to https://www.consul.io/intro/getting-started/install.html.
3. Execute the following command:

  $ consul agent -data-dir ${DATA_FOLDER} -ui -advertise 127.0.0.1 -server -bootstrap-expect 1

  ${DATA_FOLDER} could be any folder to put the data files of Consul, and it needs the read/write permission.

4. Verify the result: http://localhost:8500/ui
