#!/bin/bash -e

# convert cmdline to string array
ARGV=($@)

# grab binary path
BINPATH="${ARGV[0]}"

# binary name == service name/key
SERVICE=$(basename "$BINPATH")
ENV_FILE="$SNAP_DATA/config/$SERVICE/res/$SERVICE.env"
TAG="edgex-$SERVICE."$(basename "$0")

if [ -f "$ENV_FILE" ]; then
    logger --tag=$TAG "sourcing $ENV_FILE"
    set -o allexport
    source "$ENV_FILE" set
    set +o allexport 
fi

exec "$@"
