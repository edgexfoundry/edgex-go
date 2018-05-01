#######
Logging
#######

.. image:: EdgeX_SupportingServicesLogging.png

============
Introduction
============

Logging is critical for all modern software applications. Proper logging provides the users with the following benefits:

* Able to monitor and understand what systems are doing
* Able to understand how services interact with each other
* Problems are detected and fixed quickly
* Performance is improved

The graphic shows the high-level design architecture of EdgeX Foundry including the Logging Service.

===========================
Minimum Product Feature Set
===========================

1. Provides a RESTful API for other microservices to request log entries with the following characteristics:

* The RESTful calls should be non-blocking—meaning calling services should fire logging requests without waiting for any response from the log service—to achieve minimal impact to the speed and performance to the services.
* Support multiple logging levels, for example trace, debug, info, warn, error, fatal, and so forth.
* Support the log entry tagging; tags can be anything dictated by the calling services.
* Each log entry should be associated with its originating service.

2. Provide RESTful APIs to query, clear, or prune log entries based on any combination of following parameters:

* Timestamp from
* Timestamp to
* Log level
* Tag
* Originating service

3. Log entries should be persisted in either file or database, and the persistence storage should be managed at configurable levels
4. Take advantage of an existing logging framework internally and provide the “wrapper” for use by EdgeX Foundry
5. Follow applicable standards for logging where possible and not onerous to use on the gateway

==============================
High Level Design Architecture
==============================

.. image:: EdgeX_SupportingServicesLoggingArchitecture.png

The above diagram shows the high-level architecture for EdgeX Foundry Logging Service, which uses the Spring Boot Application Framework. Other microservices interact with EdgeX Foundry Logging Service through RESTful APIs to submit their logging requests, query historical logging, and remove historical logging. Internally, EdgeX Foundry Logging Service utilizes LOGBack as its underneath logging framework. Two configurable persistence options exist supported by EdgeX Foundry Logging Service: file or MongoDB. 

==========
Data Model
==========

.. image:: EdgeX_SupportingServicesLoggingDataModel.png


===============
Data Dictionary
===============

+---------------------+--------------------------------------------------------------------------------------------+
|   **Class Name**    |   **Descrption**                                                                           | 
+=====================+============================================================================================+
| LogEntry            | The object describing a particular log message including origin, severity, and content.    | 
+---------------------+--------------------------------------------------------------------------------------------+
| MatchCriteria       | The object describing the search parameters for a log query.                               | 
+---------------------+--------------------------------------------------------------------------------------------+

===============================
High Level Interaction Diagrams
===============================

This section shows the sequence diagrams for EdgeX Foundry Logging Service.

**Sequence Diagram for Logging Request**

.. image:: EdgeX_SupportingServicesLoggingRequest.png

**Sequence Diagram for Query Historical Logging**

.. image:: EdgeX_SupportingServicesQueryHistoricalLogging.png

**Sequence Diagram for Removing Historical Logging**

.. image:: EdgeX_SupportingServicesRemoveHistoricalLogging.png

========================
Configuration Properties
========================

The default configuration file is in the /src/main/resources folder of source code.  When interacting with the Configuration Management microservice, the configuration is in the /config/support-logging name space in Consul Key/Value Store.


+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
|   **Configuration**                                     |   **Default Value**                 |  **Dependencies**                                                         |
+=========================================================+=====================================+===========================================================================+
| read.max.limit                                          | 100                             \*  | Read data limit per invocation                                            |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| heart.beat.time                                         | 300000                          \*  | Heart beat time in milliseconds                                           |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| heart.beat.msg                                          | Logging Service heart beat      \*  | Heart beat message                                                        |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| server.port                                             | 48061                          \**  | Micro service port number                                                 |  
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| spring.cloud.consul.discovery.healthCheckPath           | /api/v1/ping                   \**  | Health checking path for Service Registry                                 | 
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| logging.persistence                                     | file                            \*  | "file" to save logging in file;                                           |
|                                                         |                                     | "mongodb" to save logging in MongoDB                                      |  
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Following config only take effect when logging.persistence=file                                                                                                           | 
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| logging.persistence.file                                |                                     | File path to save logging entries                                         |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| logging.persistence.file.maxsize                        | 10MB                            \*  | Threshold to roll and archive the logging file. It can be specified in    |
|                                                         |                                     | bytes, kilobytes, megabytes or gigabytes by suffixing a numeric value     |
|                                                         |                                     | with KB, MB and respectively GB. For example, 5000000, 5000KB, 5MB and    |
|                                                         |                                     | 2GB are all valid values, with the first three being equivalent.          |                               
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| Following config only take effect when logging.persistence=mongodb                                                                                                        |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| spring.data.mongodb.username                            | logging                        \**  | MongoDB user name                                                         |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| spring.data.mongodb.password                            | password                        \*  | MongoDB password                                                          |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| spring.data.mongodb.host                                | localhost                      \**  | MongoDB host name                                                         |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| spring.data.mongodb.port                                | 27017                          \**  | MongoDB port number                                                       |
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+
| spring.data.mongodb.database                            | logging                         \*  | MongoDB database name                                                     | 
+---------------------------------------------------------+-------------------------------------+---------------------------------------------------------------------------+

| \*means the configuration value can be changed if necessary.
| \**means the configuration value has to be replaced.
| \***means the configuration value should NOT be changed.


=======================================
Logging Service Client Library for Java
=======================================

As most of EdgeX Foundry microservices are implemented in Java, we provide a Client Library for Java, so that Java-based microservices could directly switch their Loggers to use EdgeX Foundry Logging Service.  The next graphic shows the high-level design architecture for the Java Client Library.

.. image:: EdgeX_SupportingServicesLoggingClientLibrary.png

For a Java-based microservice, follow 4 steps to use Client Library for Java:

1. Add support-logging-client as the maven dependency in pom.xml  
2. Switch your local logger to org.edgexfoundry.support.logging.client.EdgeXLogger

.. image:: EdgeX_SupportingServicesLoggingJavaLibrary1.png

3. Add mandatory configuration into properties,  e.g.  src/main/resources/application.properties,  src/test/resources/application.properties,  config folders(docker and non-docker) of config-seed project, application.properties under docker-* Bitbucket repositories 

.. image:: EdgeX_SupportingServicesLoggingJavaLibrary2.png

4. As logging-client would pick up "spring.application.name" as originService when submitting remote logging request, make sure you add proper name for such property; otherwise, logging-client would use "unknown" as default value.

.. image:: EdgeX_SupportingServicesLoggingJavaLibrary3.png

Your application will need an SLF4J implementation.  If you are using Spring Boot as part of your project, this automatically brings in an SLF4J implementation into project.  In fact, you may find multiple implementations are brought into the project and you will have to use <exclusion> elements into the pom.xml to constrain the implementations used by the project.  See core-metadata's pom.xml for an example.  In the case where your project has no implementation, you will need to add one to the pom.xml in addition to the support-logging-client.  So, for example, if you create a simple Maven project (using no other frameworks/libraries other than support-logging-client) then you will also need to add some minimal SLF4J implementation.  Here is a simple set of dependencies to achieve a working logging implementation using support-logging-client.

::

   <properties>
   	   <support-logging-client.version>1.0.0-SNAPSHOT</support-logging-client.version>
   </properties>

   <dependencies>
      	   <dependency>
		   <groupId>org.edgexfoundry</groupId>
		   <artifactId>support-logging-client</artifactId>
		   <version>${support-logging-client.version}</version>
	   </dependency>
	   <dependency>
		   <groupId>org.slf4j</groupId>
		   <artifactId>slf4j-simple</artifactId>
		   <version>1.8.0-alpha2</version>
	   </dependency>
   </dependencies>

Without the SLF4J implementation (in this case slf4j-simple), you will see errors like that below:

::

  SLF4J: Failed to load class "org.slf4j.impl.StaticLoggerBinder".
  SLF4J: Defaulting to no-operation (NOP) logger implementation
  SLF4J: See http://www.slf4j.org/codes.html#StaticLoggerBinder for further details.








