#!/bin/sh
#
# This script includes some configuration copied from the
# core-config-seed's Dockerfile, and is otherwise based
# on two shell scripts which exist in the same directory.
#
#  - launch-consul-config.sh
#  - docker-entrypoint.sh
#
set -e

CONSUL_ARGS="-server -client=0.0.0.0 -bind=127.0.0.1 -bootstrap -ui"

CONSUL_DATA_DIR=$SNAP_DATA/consul/data
CONSUL_CONFIG_DIR=$SNAP_DATA/consul/config
LOG_DIR=$SNAP_COMMON

# Handle directory creation & data cleanup
if [ -e $CONSUL_DATA_DIR ] ; then
    rm -rf $CONSUL_DATA_DIR/*
else
    mkdir -p $CONSUL_DATA_DIR
fi

if [ ! -e $CONSUL_CONFIG_DIR ] ; then
    mkdir -p $CONSUL_CONFIG_DIR    
fi

if [ ! -e $LOG_DIR ] ; then
    mkdir -p $LOG_DIR    
fi

# Run available startup hooks to have a point to store custom
# logic outside of this script. More of the things from above
# should be moved into these.
#for hook in $SNAP/startup-hooks/* ; do
#    [ -x "$hook" ] && /bin/sh -x "$hook"
#done

# TODO: remove trailing '&' if/when services are actually
# enabled in snap/snapcraft.yaml
exec $SNAP/bin/consul agent \
     -data-dir="$CONSUL_DATA_DIR" \
     -config-dir="$CONSUL_CONFIG_DIR" \
     $CONSUL_ARGS | tee $LOG_DIR/core-consul.log &
     





