#!/bin/bash -e

# the kong wrapper script from $SNAP
export KONG_SNAP="${SNAP}/bin/kong-wrapper.sh"

export num_tries=0
export MAX_KONG_UP_TRIES=10

# run kong migrations up to bootstrap the cassandra database
# note that sometimes cassandra can be in a "starting up" state, etc.
# and in this case we should just loop and keep trying
until $KONG_SNAP migrations up --yes --conf $KONG_CONF; do
    sleep 10
    # increment number of tries
    num_tries=$((num_tries+1))
    if (( num_tries > MAX_KONG_UP_TRIES )); then
        echo "max tries attempting to bring up kong"
        exit 1
    fi
done

# now start kong normally
$KONG_SNAP start --conf $KONG_CONF
