## v2 Package

[TOC]

### Overview

This package encapsulates the APIv2 implementation.

The implementation is based on [the presentation](https://zoom.us/rec/share/98x2E4vXxEFLZtbt7gLHV4oCEafMaaa803VP-vQFmk9FjGQVZvvBa-EniNhqLmEt) ([slides](https://wiki.edgexfoundry.org/download/attachments/329472/APIV2_20200123.pdf?version=1&modificationDate=1579821274000&api=v2)) given to the Core Working Group on January 23, 2020.

#### Disclaimer

This document and the implementation it describes is a work in progress and subject to change.  The architectural approach being taken is evolutionary; the existing implementation does not currently support all aspects of the final design.



### Approach

The scope of APIv2 prevents complete implementation across all services within a single EdgeX release cycle.  The community has decided to release Geneva with APIv1 and "experimental/beta" APIv2 support for  some subset of services.  The intention is for the future EdgeX 2.0 release to include only APIv2 endpoints.

APIv2 is being treated as greenfield development.  This reduces risk by minimizing overlap with APIv1 implementation.  It ensures APIv2 is able to be developed and evolve separately from APIv1.  It also provides a clear delineation between APIv1 and APIv2 and makes it easier (and less likely to introduce last-minute issues) to remove the APIv1 functionality for the future EdgeX 2.0 release.

To this end, the community decided that APIv2 persistence will be separate from APIv1 persistence.  This supports independent schema development at the expense of data access across the two API versions.  That is, data persisted via APIv1 will not be available via APIv2 (and vice-versa).



### Definitions

- ***Use-case:*** discrete behavior initiated by external interaction.  The external interaction could be a user or another service.  A use case is associated with a version, a type, and an action.
- ***Version:*** the major value of the semantic version of a specific use- case implementation (e.g. "2").
- ***Type:*** a mnemonic associated with a specific use-case implementation (e.g. "ping").
- ***Action:*** a generic (HTTP method-equivalent) name associated with a specific use-case implementation; currently defined are "create", "read", "update", "delete", and "command."
- ***DTO:*** data transfer object.  An object with no behavior used only for data transmission; e.g. the request and response objects that define the structure and content of the data required by and returned by a use-case.
- ***Use-case Request:*** a request DTO instance that maps to a single invocation of a use-case.
- ***Use-case Response:*** a response DTO instance returned from a single invocation of a use-case. 
- ***Transport Request:***  in the current implementation, this is a single HTTP request.  A single transport request can contain one or more use-case requests and an equivalent number of use-case responses.
- ***Use-case Endpoint:*** URL that accepts one or more use-case-specific requests and returns an equivalent number of use-case responses.  
- ***Batch Endpoint:*** un-versioned URL (`/api/batch`) that accepts one or more use-case requests and returns an equivalent number of use-case responses. 



### Layered Architecture

The implementation is based on the layers common to a domain-driven design application that uses traditional Layers Architecture (see Buschmann, Frank, et al. 1996. *Pattern-Oriented Software Architecture, Volume 1: A System of Patterns*. New York: Wiley).

The implementation spans 4 architectural layers:

- ***User Interface Layer***  
  - Transport-specific implementation.  
  - Controller (from [model-view-controller](https://en.wikipedia.org/wiki/Model%E2%80%93view%E2%80%93controller)) implementation.

- ***Application Layer***
  - Use-case implementation.
  - DTO definitions.
  - DTO validation implementation.
  - Repository contract definitions.
- ***Domain Layer***
  - Domain model implementation; e.g "business logic."
  - Domain validation implementation.
  - Domain service implementation.
- ***Infrastructure Layer***
  - Repository implementation.
  - Ancillary supporting structural implementation.

Layers are ordered as they appear (from high -> low): User Interface -> Application -> Domain -> Infrastructure.

The implementation is a relaxed layered architecture.  Higher layers can depend on lower layers.  Lower layers may not depend on higher layers.



### Project Structure

The intention is to place all APIv2-related code in the `internal/pkg/v2` sub-tree:

```
internal/pkg/v2						APIv2-related implementation sub-tree
	README.md							This document

internal/pkg/v2/application			Application Layer
	delegate							Executable contract delegate implementation
	dto/v2dot0							Version 2.0 DTO definitions
	usecases							Use-case implementations
		common								Common cross-service use-cases
	validator							DTO validation implementation

internal/pkg/v2/domain				Domain Layer

internal/pkg/v2/infrastructure		Infrastructure Layer
	test								APIv2-related acceptance test support

internal/pkg/v2/ui					User Interface Layer
	common								Transport-independent implementation
		batchdto							Batch envelope DTO definitions
		middleware/debugging				Example debugging middleware 
		routable							Routable delegate implementation
	http								HTTP transport-specific implementation
		api									Controller implementation					
			common								Common cross-service controllers/tests
			core								Core-service controllers/tests
			support								Support-service controllers/tests
			system								System Management controllers/tests
		handle								Common request handling implementation
		routing								Version, kind, action router
```

Note the separation of common cross-service code from service-specific code.



### Use-case Endpoints

A use-case endpoint is versioned; e.g. `/api/v2/metrics`.

A use-case endpoint accepts either a single use-case request or an array of use-case requests.

A single use-case request returns a single corresponding use-case response.  Success is indicated by an 200 HTTP status code.  The use-case response contains a transport-independent result code that indicates success or a specific failure reason.

An array of use-case requests returns an array of corresponding use-case responses.  A 207 HTTP status code will be returned regardless of the success or failure of the individual requests.  Each use-case response contains a transport-independent result code that indicates success or a specific failure reason.

Multiple use-case requests received in a single transport request must all be of same type (which corresponds to the request DTO expected by the underlying use-case).

Multiple use-case requests received in a single transport request are processed concurrently.  Results are aggregated and returned as a single array of responses in the transport's response.  The relative position of a use-case response in the response's array has no correlation to the relative position of the corresponding use-case request in the request's array.



### Batch Endpoint

There is only one batch endpoint and it is unversioned: `/api/batch`.  

A RequestEnvelope DTO and a ResponseEnvelope DTO wrap use-case request DTOs and use-case response DTOs respectively for requests and response sent via the batch endpoint.  These batch-specific DTOs contain version, type, and action properties.  They also include a strategy and a content property.  The content property contains a use-case-specific request or response that corresponds to the version, type, and action in the envelope.

A batch endpoint accepts an array of wrapped use-case requests and returns an array of wrapped corresponding use-case responses.  A 207 HTTP status code will be returned regardless of the success or failure of the individual requests.  Each use-case response contains a transport-independent result code that indicates success or a specific failure reason.  

Multiple wrapped use-case requests received in a single transport request may be of mixed types.

Multiple wrapped use-case requests received in a single transport request are processed sequentially.  Results are aggregated and returned as a single array of wrapped responses in the transport's response. The relative position of a wrapped use-case response in the response's array has no correlation to the relative position of the corresponding wrapped use-case request in the request's array.

In the current implementation, batch endpoint support for a specific use-case is effectively free if you follow the patterns established by the initial implementation (i.e. create a controller and use-case around a unique version, kind, and action).



### Use-cases

Use-cases live in the application layer. 

A single use-case can handle one or more version, type, and action variations.

A use-case typically has its own specific request DTO (for input) and response DTO (for output).

A use-case is implemented following the [Gang of Four Command design pattern](https://en.wikipedia.org/wiki/Command_pattern) and implements the `Routable` contract.



### DTO Definitions

Request and Response DTOs are defined in the application layer.  To facilitate acceptance testing (see elsewhere in this document), all DTO variations for all minor releases are maintained in a versioned directory structure (see `internal/pkg/v2/application/dto`).



### DTO Validation

DTO validation lives in the application layer; specifically in `internal/pkg/v2/application/validator`.

Validation is an orthogonal concern and occurs outside of use-case implementation.

Validators are applied to use-cases during wire-up.  Validators can be layered.

Validation shares implementation with middleware; the underlying implementation to [curry](https://en.wikipedia.org/wiki/Currying) `application.Executable` behavior is common for both.



### Router

The router lives in the user interface layer.  

It maps version, type, and action to a `Routable` contract implementation.



### Controllers

Controllers live in the user interface layer and implement the `Controller` contract.

They bridge the user interface (in the current implementation, this is an endpoint invocation) to a specific use-case implementation.



### Handlers

Two generic handler implementations are provided -- one for all use-case endpoint invocations and one for batch endpoint invocations.  

These generic implementations provide generic marshalling and unmarshalling of JSON to and from Golang structures.  They use the Router to determine what Use-case implementation to delegate to.



### Middleware

Middleware lives in the user interface layer.

It typically contains behavior executed for each use-case request.  It supports both pre- and post-use-case processing.  It can modify the request (as it heads towards the use-case for processing) and/or the response (as it heads towards the user after use-case processing has completed).

Middleware is optional and can be selectively enabled.  It can be added to all use-cases for universal orthogonal concerns (the debugging middleware provides a good example of this).  It can also be selectively applied to individual use-cases.

Middleware is extensible and can be layered.  You can apply different middleware implementations to the same use-case; each middleware is implemented and executed in isolation from every other middleware.

Middleware shares implementation with DTO validation; the underlying implementation to [curry](https://en.wikipedia.org/wiki/Currying) `application.Executable` behavior is common for both.



#### Example Debugging Middleware

A debugging middleware is provided as an example middleware implementation.  It is selectively enabled for all use-case requests by adding the `-debug` command line flag when executing a service.  When enabled, the processing time (in milliseconds) as well as the version, kind, action, request content, and response content properties are all logged for every use-case request.



### Common Behavior

Certain services share certain behavior -- for example, the ping, metrics, configuration, and version endpoints are commonly supported across all core, support, and system management services.

In this design, there is a single implementation of each of these behaviors.  That single implementation is leveraged by each service that requires it.  Change that single implementation, rebuild all of the core, support, and system management services, and they'll all have the updated behavior.



### mux.Router Wire Up

APIv2 routes are added to the mux.Router in each service's bootstrap handler.  For core, support, and system management services, this is currently in the service's init.go.  Actual implementation has been encapsulated by the `loadV2Routes` method. 



### Acceptance Tests

Generic support for golang-based acceptance tests has been implemented by `internal/pkg/v2/infrastructure/test`; see the package's README.md for additional detail.

Tests instantiate an instance of the service being tested inside the test runner context.  The intention is to leverage an in-process in-memory implementation of the persistence contract.

Acceptance tests are written to exercise all HTTP methods on a given endpoint.  This ensures both valid and invalid HTTP methods return the expected response.

Each common behavior (as described elsewhere in this document; e.g ping, metrics, configuration, and version endpoints) has a common test implementation.  Each service implementing the common behavior has its own test implementation that instantiates the service and delegates the actual testing to the common test implementation.

A generic test exists to verify concurrent execution of multiple use-case request made on a single use-case endpoint.  A generic test exists to verify sequential execution of multiple requests made on the batch endpoint.  When writing acceptance tests for new use-cases, the manner of processing does not need to be tested.

Acceptance tests are written to ensure backward- and forward-compatibility across minor API versions.  This is done by retaining each minor release's DTOs, using them to systematically execute requests against the latest implementation, and ensuring the response returned conforms to expectations.

As implemented, tests make heavy use of constants and contain no hardcoded JSON.

#### New Required `REPO_ROOT` Environment Variable

Since the new acceptance tests instantiate an instance of the service being tested inside the test runner context, the require access to the default configuration.toml files.  These service-specific files are defined in the various service-specific sub-directories under `/cmd`. Unfortunately, the way we execute `go test` in the builds means the directory containing the test becomes the working directory. 

The newly added `REPO_ROOT` environment variable should have its value set to the absolute path of the root of the local copy of the edgex-go repository.  This provides a common point for the tests to specify the appropriate directory containing a service's configuration files thereby allowing the service to successfully bootstrap itself.

This environment variable (and an appropriate value) will need to be added to any environment that executes `go test`.

 

### Forward Flexibility In Today's Design

#### Transport substitution

The design supports the ability to add and/or substitute transports (e.g. gRPC, sockets, websockets, pub-sub, asynchronous messaging, etc.).

#### Single Configuration-Driven Executable

The design supports a single executable capable of supporting endpoints from one or more of today's existing services.  This could be hardcoded -- as it is in the current Fuji release.  It could also easily be configuration-driven -- a single executable could provide any or all existing functionality for all core, support, and system management services based on a configuration file.  This would simplify artifact creation and provide a more flexible deployment story.

#### RequestEnvelope's Strategy Property

The RequestEnvelope includes a strategy property; currently defined are "sync", "async-push", and "asynch-poll".  The strategy property is intended to provide a transport-agnostic way (via an extensible [Strategy pattern](https://en.wikipedia.org/wiki/Strategy_pattern) implementation) of specifying the desired processing strategy to be applied to a use-case request.

The "sync" option is implemented by default.  It processes the use-case request synchronously and returns a result only when processing has been completed.

The "asynch-push" option provides for asynchronous processing of a use-case request.  The caller supplies an address as part of the request.  The request is accepted for processing and an intermediate result is immediately returned.  When request processing has asynchronously completed, the result will be pushed to the address provided.  This option has not been implemented and requires further definition and refinement.

The "async-poll" option provides for asynchronous processing of a use-case request.  The request is accepted for processing and a unique token is immediately returned.  When request processing has asynchronously completed, the result will be stored for some finite timeframe.  The caller would provide the unique token to query for the result through a yet to be defined endpoint/process.  This option has not been implemented and requires further definition and refinement.

#### Alternate Processing Strategies for Batch Requests

It's currently possible (and in the batch endpoint's case it's required) to supply an array of use-case requests in a transport request.

There's forward vision that would provide for wrapping an array of use-case requests in an new object that would include a strategy field for processing.  The strategy field would specify a specific processing strategy (via an extensible [Strategy pattern](https://en.wikipedia.org/wiki/Strategy_pattern) implementation).  

This would allow for complex saga-like interaction (i.e. validate that the array contains all of the items necessary to do some complex behavior; process all of the objects of type A concurrently; wait for type A processing to complete; process type B, type C, and type D objects sequentially and in that order; return a result). 

This would also allow for a strategy that provided for simple request-ignorant concurrent processing of all RequestEnvelope DTOs.

Although this concept could be implemented generically for use-case endpoints, it's only currently envisioned for the batch endpoint (as the batch endpoint provides the entire APIv2 urcontract expected to be implemented by other transport implementations).



### Known Open Issues Still To Be Addressed

1. An example of a service-specific endpoint needs to be implemented.  This will illustrate how controllers for HTTP GET and DELETE method-based endpoints will create use-case request DTOs from HTTP headers and/or URL content.  It will also illustrate the approach to persistence, domain models, and how persistence schema and DTO definition diverge.
2. The APIv2 specifications are inconsistent.  Certain definitions related to the ping, metrics, configuration, and version endpoints are accurate.  The rest of the document is not.  I intend to update the documents to make them consistent after the service-specific endpoint example has been completed, reviewed, and merged.
3. Creation of an ADR to document the design, architecture, and approach to the APIv2 implementation.  I intend to repurpose a good portion of this document to create the ADR after the service-specific endpoint example has been completed, reviewed, and merged.
4. go-mod-core-contracts integration.
5. Further definition and strategy pattern-based implementation of the RequestEnvelope's strategy property.