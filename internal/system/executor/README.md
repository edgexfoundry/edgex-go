# EdgeX Foundry System Management Executor
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

This README.md is geared toward a developer interested in creating their own executor. It includes related information 
    that ties in with the System Management Agent (aka SMA). The main points are:

- How the SMA passes service name, and action on the command line.
- Current proxy-like behavior for stop/start/restart operations -- the SMA passes parameters received to executor as-is.
- The Metrics Result Contract (and its support for embedding executor-specific results).
- Expected format of operation result (based upon the existing Docker executor implementation).

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
{"operation":"stop","service":"edgex-support-notifications","executor":"docker","Success":true}
```
- ./sys-mgmt-executor edgex-support-notifications start
```
bash-5.0# ./sys-mgmt-executor edgex-core-data start
{"operation":"start","service":"edgex-core-data","executor":"docker","Success":true}
```
- ./sys-mgmt-executor edgex-support-notifications restart
```
bash-5.0# ./sys-mgmt-executor edgex-support-notifications restart
{"operation":"restart","service":"edgex-support-notifications","executor":"docker","Success":true}
```

# Current Proxy-like Behavior for Stop/Start/Restart Operations #

## The three POST operations delegated to the Executor ##
- These operations involves a POST operation each, where the body is provided as a JSON payload.
- Here is an example payload for the "start" operation:
```
{
   "action":"start",
   "services":[
      "edgex-core-command",
      "edgex-core-data",
      "edgex-core-metadata"
   ]
}
```
- And the Accompanying response:
```
[
    {
        "Success": true,
        "executor": "docker",
        "operation": "start",
        "service": "edgex-core-command"
    },
    {
        "Success": true,
        "executor": "docker",
        "operation": "start",
        "service": "edgex-core-data"
    },
    {
        "Success": true,
        "executor": "docker",
        "operation": "start",
        "service": "edgex-core-metadata"
    }    
]
```
- And here is an example payload for the "restart" operation:
```
{
   "action":"restart",
   "services":[
      "edgex-support-notifications"
   ]
}
```
- And the accompanying response:
```
[
    {
        "Success": true,
        "executor": "docker",
        "operation": "restart",
        "service": "edgex-support-notifications"
    }
]
```

### Note ###
- The example payload for the "stop" operation, while not shown here, is similar to the ones shown above.
- The SMA passes parameters received to the executor as-is and that the parameter `"executor": "docker"` points out the 
    executor implementation being used.  Each executor implementation should return its own unique value for the 
    `executor` field.

# Metrics Result Contract #

- Metrics result contract (and its support for embedding executor-specific results).
Expected format of operation result (based upon the existing Docker executor implementation).


### Note ###
- The metrics result contract stipulates the following fields, highlighted inline in this example response: 
```
[
    {
        "Success": boolean,                             // Required: True or false
        "executor": "docker",                           // Required: The (reference) executor implementation
        "operation": "metrics",                         // Required: The operation
        "result": {            
            "cpuUsedPercent": 5.08,                     // Required: CPU Usage
            "memoryUsed": 5488247,                      // Required: Memory Usage
            "raw": {                                    // Optional, executor-specific results
                "block_io": "8.19kB / 0B",              // Optional, executor-specific results
                "cpu_perc": "5.08%",                    // Optional, executor-specific results
                "mem_perc": "0.26%",                    // Optional, executor-specific results
                "mem_usage": "5.234MiB / 1.952GiB",     // Optional, executor-specific results
                "net_io": "277kB / 194kB",              // Optional, executor-specific results
                "pids": "14"                            // Optional, executor-specific results
            }                                           // Optional, executor-specific results
        },
        "service": "edgex-core-command"                 // Required: Name of service whose metrics were fetched
    },
    {
        "Success": true,
        "executor": "docker",
        "operation": "metrics",
        "result": {
            "cpuUsedPercent": 5.33,
            "memoryUsed": 5373952,
            "raw": {
                "block_io": "143kB / 0B",
                "cpu_perc": "5.33%",
                "mem_perc": "0.26%",
                "mem_usage": "5.125MiB / 1.952GiB",
                "net_io": "130kB / 118kB",
                "pids": "13"
            }
        },
        "service": "edgex-support-notifications"
    }
]
```
- As highlighted above in the results section (for the first of the two services whose metrics were fetched), a handful 
    of fields aer stipulated to be part of the metrics result contract. 
- The remaining fields fall into the category of executor's support for embedding executor-specific results.
- The user has the choice of whether to further process the embedded _executor_-specific results.

### Note ###
- The expected format of the result is based upon the current _Docker_ executor implementation (as highlighted by the 
    JSON key-value pair _"executor": "docker"_).

## License
[Apache-2.0](LICENSE)

