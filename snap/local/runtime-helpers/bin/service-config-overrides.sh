#!/bin/bash -e

# convert cmdline to string array
ARGV=($@)

# grab binary path
BINPATH="${ARGV[0]}"

# binary name == service name/key
SERVICE=$(basename "$BINPATH")
SERVICE_ENV="$SNAP_DATA/config/$SERVICE/res/$SERVICE.env"

if [ -f "$SERVICE_ENV" ]; then
    logger "edgex service override: : sourcing $SERVICE_ENV"
    source "$SERVICE_ENV"
fi

exec "$@"
