#!/bin/bash
# note: -e flag is not used in this one-shot service
# we don't want to exit out the whole Consul process when ACL bootstrapping failed, just that
# Consul won't have ACL to be used

# TODO: determine if this script is really necessary. I'm not sure why the command below isn't
# simply specified as the service command, with the command-chain service-overrides script to
# handle sourcing the .env file

SERVICE_ENV="$SNAP_DATA/config/security-bootstrapper/res/security-bootstrapper.env"
if [ -f "$SERVICE_ENV" ]; then
    logger "edgex service override: : sourcing $SERVICE_ENV"
    source "$SERVICE_ENV"
fi

# setup Consul's ACL via security-bootstrapper's subcommand
"$SNAP"/bin/security-bootstrapper -confdir "$SNAP_DATA"/config/security-bootstrapper/res setupRegistryACL
setupACL_code=$?
if [ "${setupACL_code}" -ne 0 ]; then
  echo "$(date) failed to set up Consul ACL"
fi
