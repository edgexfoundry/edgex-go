#!/bin/bash

# unfortunately until snapd bug https://bugs.launchpad.net/snapd/+bug/1796125 
# is fixed, on install services startup is not guaranteed in the right order
# and as such, we need to do a little bit of hand holding for the go services,
# which only have 10 seconds for consul to come alive, but on some systems
# it can be more than 10 seconds until consul comes alive in which case
# the services will fail to startup 
# this gives consul 50 seconds to come up
MAX_TRIES=10

while [ "$MAX_TRIES" -gt 0 ] ; do
    CONSUL_RUNNING=$(curl -s http://localhost:8500/v1/catalog/service/consul)

    if [ $? -ne 0 ] ||
       [ -z "$CONSUL_RUNNING" ] ||
       [ "$CONSUL_RUNNING" = "[]" ] || [ "$CONSUL_RUNNING" = "" ]; then
        echo "$1: consul not running; remaing tries: $MAX_TRIES"
        sleep 5
        MAX_TRIES=$(($MAX_TRIES - 1))
    else
	    break
    fi
done
