#!/bin/bash -e

# start consul in the background
"$SNAP/bin/consul" agent \
    -data-dir="$SNAP_DATA/consul/data" \
    -config-dir="$SNAP_DATA/consul/config" \
    -server -client=0.0.0.0 -bind=127.0.0.1 -bootstrap -ui &

# loop trying to connect to consul, as soon as we are successful exit
# NOTE: ideally consul would be able to notify systemd directly, but currently 
# it only uses systemd's notify socket if consul is _joining_ another cluster
# and not when bootstrapping
# see https://github.com/hashicorp/consul/issues/4380

# to actually test if consul is ready, we simply check to see if consul 
# itself shows up in it's service catalog
# also note we don't have a timeout here because we use start-timeout for this
# daemon so systemd will kill us if we take too long waiting for this
CONSUL_URL=http://localhost:8500/v1/catalog/service/consul
until [ -n "$(curl -s $CONSUL_URL | jq -r '. | length')" ] && 
    [ "$(curl -s $CONSUL_URL | jq -r '. | length')" -gt "0" ] ; do
    sleep 1
done
