#!/bin/sh -e

# check if security-secret-store is on/off
# if it's not specified assume it's on

SEC_STORE=$(snapctl get security-secret-store)
if [ "$SEC_STORE" = "off" ]; then
    # then export the env var as false to turn everything off
    EDGEX_SECURITY_SECRET_STORE=false
    export EDGEX_SECURITY_SECRET_STORE
fi

exec "$@"
