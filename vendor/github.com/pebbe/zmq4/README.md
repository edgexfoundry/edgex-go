A Go interface to [ZeroMQ](http://www.zeromq.org/) version 4.


----------------------------------------------------------------

## Warning

Starting with Go 1.14, on Unix-like systems, you will get a lot of
interrupted signal calls. See the top of a package documentation
for a fix.

----------------------------------------------------------------


[![Go Report Card](https://goreportcard.com/badge/github.com/pebbe/zmq4)](https://goreportcard.com/report/github.com/pebbe/zmq4)
[![GoDoc](https://godoc.org/github.com/pebbe/zmq4?status.svg)](https://godoc.org/github.com/pebbe/zmq4)

This requires ZeroMQ version 4.0.1 or above. To use CURVE security in
versions prior to 4.2, ZeroMQ must be installed with
[libsodium](https://github.com/jedisct1/libsodium) enabled.

Partial support for ZeroMQ 4.2 DRAFT is available in the alternate
version of zmq4 `draft`. The API pertaining to this is subject to
change. To use this:

    import (
        zmq "github.com/pebbe/zmq4/draft"
    )

For ZeroMQ version 3, see: http://github.com/pebbe/zmq3

For ZeroMQ version 2, see: http://github.com/pebbe/zmq2

Including all examples of [ØMQ - The Guide](http://zguide.zeromq.org/page:all).

Keywords: zmq, zeromq, 0mq, networks, distributed computing, message passing, fanout, pubsub, pipeline, request-reply

### See also

 * [go-zeromq/zmq4](https://github.com/go-zeromq/zmq4) — A pure-Go implementation of ØMQ (ZeroMQ), version 4
 * [go-nanomsg](https://github.com/op/go-nanomsg) — Language bindings for nanomsg in Go
 * [goczmq](https://github.com/zeromq/goczmq) — A Go interface to CZMQ
 * [Mangos](https://github.com/go-mangos/mangos) — An implementation in pure Go of the SP ("Scalable Protocols") protocols

## Requirements

zmq4 is just a wrapper for the ZeroMQ library. It doesn't include the
library itself. So you need to have ZeroMQ installed, including its
development files. On Linux and Darwin you can check this with (`$` is
the command prompt):

```
$ pkg-config --modversion libzmq
4.3.1
```

The Go compiler must be able to compile C code. You can check this
with:
```
$ go env CGO_ENABLED
1
```

You can't do cross-compilation. That would disable C.

## Install

    go get github.com/pebbe/zmq4

## Docs

 * [package help](http://godoc.org/github.com/pebbe/zmq4)
 * [wiki](https://github.com/pebbe/zmq4/wiki)

## API change

There has been an API change in commit
0bc5ab465849847b0556295d9a2023295c4d169e of 2014-06-27, 10:17:55 UTC
in the functions `AuthAllow` and `AuthDeny`.

Old:

    func AuthAllow(addresses ...string)
    func AuthDeny(addresses ...string)

New:

    func AuthAllow(domain string, addresses ...string)
    func AuthDeny(domain string, addresses ...string)

If `domain` can be parsed as an IP address, it will be interpreted as
such, and it and all remaining addresses are added to all domains.

So this should still work as before:

    zmq.AuthAllow("127.0.0.1", "123.123.123.123")

But this won't compile:

    a := []string{"127.0.0.1", "123.123.123.123"}
    zmq.AuthAllow(a...)

And needs to be rewritten as:

    a := []string{"127.0.0.1", "123.123.123.123"}
    zmq.AuthAllow("*", a...)

Furthermore, an address can now be a single IP address, as well as an IP
address and mask in CIDR notation, e.g. "123.123.123.0/24".
