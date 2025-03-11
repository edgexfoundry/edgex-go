# Ziti Metrics Library

This is a metrics library which is built on, and extends the
[go-metrics](https://github.com/rcrowley/go-metrics) library.

It extends it by adding the following:

1. Support for interval counters. These collect event counts related to a given identifier, over a given interval. The interval buckets are flushed regularly. These are good for things like collecting
   usage data.
1. A Dispose method is defined on metrics, so they can be cleaned up, for metrics tied to transient entities.
1. Reference counted metrics