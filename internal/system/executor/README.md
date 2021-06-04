# EdgeX Foundry System Management Executor
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

This README.md is geared toward a developer interested in creating their own executor. It includes related information 
    that ties in with the System Management Agent (aka SMA). The main points are:

- How the SMA passes service name, and action on the command line.
- Current proxy-like behavior for stop/start/restart operations -- the SMA passes parameters received to executor as-is.
- The Metrics Result Contract (and its support for embedding executor-specific results).

# Passing Parameters to SMA on Command Line #

#### Usage ####
./sys-mgmt-executor [service-name] [operation]

Where:
- "service-name" is the name of the service to apply the operation to.
- "operation" can be one of [start, stop, restart]

**Note**: neither operation, nor service-name are verified by the SMA before passing to the executor, the executor is responsible for ensuring invalid service names and operations are handled gracefully.

#### Examples of Usage (and Sample Responses) ####
- ./sys-mgmt-executor edgex-support-notifications stop
```
bash-5.0# ./sys-mgmt-executor edgex-support-notifications stop
""
```
- ./sys-mgmt-executor edgex-support-notifications start
```
bash-5.0# ./sys-mgmt-executor edgex-core-data start
""
```
- ./sys-mgmt-executor edgex-support-notifications restart
```
bash-5.0# ./sys-mgmt-executor edgex-support-notifications restart
""
```

# Current Proxy-like Behavior for Stop/Start/Restart Operations #

## The three POST operations delegated to the Executor ##
- These operations involves a POST operation each, where the body is provided as a JSON payload.
- Here is an example payload for the "start" operation:
```
{
    [
        "apiVersion": "v2",
        "serviceName": "edge-core-command",
        "action": "start"
    ],
    [
        "apiVersion": "v2",
        "serviceName": "edge-core-metadata",
        "action": "start"
    ],
    [
        "apiVersion": "v2",
        "serviceName": "edge-core-data",
        "action": "start"
    ]
}
```
- And the Accompanying response:
```
[
    {
        "apiVersion": "v2",
        "statusCode": 200,
        "serviceName": "edgex-core-command"
    },
    {
        "apiVersion": "v2",
        "statusCode": 200,
        "serviceName": "edgex-core-metadata"
    },
    {
        "apiVersion": "v2",
        "statusCode": 200,
        "serviceName": "edgex-core-data"
    }
]
```
- And here is an example payload for the "restart" operation:
```
{
    [
        "apiVersion": "v2",
        "serviceName": "edgex-support-notifications",
        "action": "restart"
    ]
}
```
- And the accompanying response:
```
[
    {
        "apiVersion": "v2",
        "statusCode": 200,
        "serviceName": "edgex-support-notifications"
    }
]
```

### Note ###
- The example payload for the "stop" operation, while not shown here, is similar to the ones shown above.

# Metrics Result Contract #

- The metrics result allows user to provide desired metrics field in `map[string]interface{}` format, here is the example response: 
```
{
    "cpuUsedPercent": 5.08,
    "memoryUsed": 5488247,
    "raw": {
        "block_io": "8.19kB / 0B",
        "cpu_perc": "5.08%",
        "mem_perc": "0.26%",
        "mem_usage": "5.234MiB / 1.952GiB",
        "net_io": "277kB / 194kB",
        "pids": "14"
    }
}
```

### Note ###
- The expected format of the result is based upon the current _Docker_ executor implementation. 

## License
[Apache-2.0](../../../LICENSE)

