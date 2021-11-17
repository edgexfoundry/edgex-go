/*
A Go interface to ZeroMQ (zmq, 0mq) version 4.

For ZeroMQ version 3, see: http://github.com/pebbe/zmq3

For ZeroMQ version 2, see: http://github.com/pebbe/zmq2

http://www.zeromq.org/

See also the wiki: https://github.com/pebbe/zmq4/wiki

----

A note on the use of a context:

This package provides a default context. This is what will be used by
the functions without a context receiver, that create a socket or
manipulate the context. Package developers that import this package
should probably not use the default context with its associated
functions, but create their own context(s). See: type Context.

----

Since Go 1.14 you will get a lot of interrupted system calls.

See: https://golang.org/doc/go1.14#runtime

There are two options to prevent this.

The first option is to build your program with the environment variable:

    GODEBUG=asyncpreemptoff=1

The second option is to let the program retry after an interrupted system call.

Initially, this is set to true, for the global context, and for contexts
created with NewContext().

When you install a signal handler, for instance to handle Ctrl-C, you should
probably clear this option in your signal handler. For example:

    zctx, _ := zmq.NewContext()

    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        chSignal := make(chan os.Signal, 1)
        signal.Notify(chSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
        <-chSignal
        zmq4.SetRetryAfterEINTR(false)
        zctx.SetRetryAfterEINTR(false)
        cancel()
    }()

----

*/
package zmq4
