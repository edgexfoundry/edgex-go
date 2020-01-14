## Test Package

[TOC]

### Overview

This package provides support for APIv2 golang-based acceptance tests.

It is based on the presentation given to the QA/Test Working Group on December 17, 2019.



### Important Concepts

#### SUT (System Under Test)

The `NewSUT` function accepts a list of environment variables, command line arguments, a path to the service's configuration file, and a reference to the service's main entry function (which should conform to the `MainFunc` contract).

It starts the service in a separate goroutine and returns a `cancel` function (to stop the service), a `sync.WaitGroup` (to wait for the goroutine to end after `cancel` is called), and a reference to the service's `mux.Router` (to be used by the caller to execute tests).

It assumes you are testing a service based on  [`go-mod-bootstrap`](https://github.com/edgexfoundry/go-mod-bootstrap).  

#### Timer

A simple timer to measure time in milliseconds is provided.  The timer captures the current time when it is instantiated and continues until `Stop` is called.  Corresponding `AssertElapsedInsideDeviation` and `AssertElapsedInsideStandardDeviation` utility functions provide the capability to assert timing-related concerns.

#### Utility Functions

Several test-related utility functions are provided:

- `AssertContentTypeIsJSON` provides a common implementation to assert the HTTP content-type header indicates the content is JSON.
- `AssertJSONBody` provides a common implementation to assert a response body has the expected values regardless of whether the response is a single object or an array of objects.  Further, successful assertion does not rely on order when asserting an array of objects.
- `AssertElapsedInsideDeviation` provides a common implementation to assert a test timer falls within an expected range (in milliseconds).
- `AssertElapsedInsideStandardDeviation` provides a common implementation to assert a test timer falls within an standard range -- currently plus or minus 50 milliseconds.
- `FactoryRandomString`provides a simple factory function that returns a random string to provide unique data for a specific test instance's execution.
- `InvalidJSON` provides a simple factory function that returns a `[]byte` that contains invalid JSON content; i.e. content that will fail a `json.Unmarshal`call.
- `Marshal` provides a generic function to convert a golang structure into its JSON string representation (and typed as a `[]byte`).
- `SendRequest` provides a common implementation to execute an HTTP request via a specified HTTP method, to a specified URL, with specified HTTP body content via a provided `mux.Router` (which is returned by the `NewSUT` function).
- `ValidMethods` provides a comprehensive list of valid HTTP methods.
- `Join` and `Name` provides common implementations for joining strings and creating a test name from a list of variant descriptions.



### Dependencies

[`github.com/stretchr/testify/assert`](https://github.com/stretchr/testify) 

[`github.com/gorilla/mux`](https://github.com/gorilla/mux)