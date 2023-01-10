#!/bin/sh -e

# This script is used in the command-chain before starting EdgeX apps

security=$(snapctl get security)
if [ "$security" = "disabled" ]; then
    # set to false so that the app doesn't use the secret store
    export EDGEX_SECURITY_SECRET_STORE=false
fi

exec "$@"
