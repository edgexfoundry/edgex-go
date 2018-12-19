#!/bin/bash -e

# wait for consul to come up
$SNAP/bin/wait-for-consul.sh $1

# if we get here, just assume that consul is up, otherwise this 
# service will fail after 10 seconds
cd $SNAP_DATA/config/$1

# note we have to exec the process so it takes over the 
# same pid as the calling bash process since this bash script
# is forked from another script that systemd runs
# this ensures that systemd will end up tracking the actual go 
# service process and not the shell process
exec $SNAP/bin/$1 "${@:2}"
