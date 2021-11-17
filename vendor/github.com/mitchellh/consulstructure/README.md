# consulstructure

consulstructure is a Go library for decoding [Consul](https://www.consul.io)
data into a Go structure and keeping it in sync with Consul.

The primary use case for this library is to be able to build native
Consul-based configuration into your Go applications without needing
glue such as [consul-template](https://github.com/hashicorp/consul-template).

## Installation

Standard `go get`:

```
$ go get github.com/mitchellh/consulstructure
```

## Features

Below is a high-level feature list:

  * Watch a key prefix in Consul KV to populate a Go structure.

  * Notification on a channel when configuration is updated.

  * Configuration structures support all Go primitive types, maps, and structs.
    Slices and arrays aren't supported since they don't mean anything in
    the data model of Consul KV.

  * Nested and embedded structs in configuration structures work.

  * Set quiescence periods to avoid a stampede of configuration updates
    when many keys are updated in a short period of time.

  * Supports all connection features of Consul: multi-datacenter, encryption,
    and ACLs.

## Usage & Example

For docs see the [Godoc](http://godoc.org/github.com/mitchellh/consulstructure).

An example is shown below:

```go
import (
    "fmt"

    "github.com/mitchellh/consulstructure"
)

// Create a configuration struct that'll be filled by Consul.
type Config struct {
    Addr     string
    DataPath string `consul:"data_path"`
}

// Create our decoder
updateCh := make(chan interface{})
errCh := make(chan error)
decoder := &consulstructure.Decoder{
    Target:   &Config{},
    Prefix:   "services/myservice",
    UpdateCh: updateCh,
    ErrCh:    errCh,
}

// Run the decoder and wait for changes
go decoder.Run()
for {
    select {
    case v := <-updateCh:
        fmt.Printf("Updated config: %#v\n", v.(*Config))
    case err := <-errCh:
        fmt.Printf("Error: %s\n", err)
    }
}
```

## But Why Not a File?

A file is the most portable and technology agnostic way to get configuration
into an application. I'm not advocating this instead of using files for the
general case.

For organizations that have chosen Consul as their technology for configuration,
services, etc. being able to build services that can be started without
any further configuration and immediately start running is very attractive.
You no longer have a configuration step where you have to setup services
and file templates and so on with tools like
[consul-template](https://github.com/hashicorp/consul-template).

You just install the Go binary, start it, and it is going. To update it,
you just update the settings in Consul, and the application automatically
updates. No more SIGHUP necessary, no more manual restarts.
