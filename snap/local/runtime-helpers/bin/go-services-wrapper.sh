#!/bin/bash -e

systemd-notify --ready

# wait for consul to come up
$SNAP/bin/wait-for-consul.sh $1

# if we get here, just assume that consul is up, otherwise this 
# service will fail after 10 seconds
cd $SNAP_DATA/config/$1
$SNAP/bin/$1 "${@:2}"
