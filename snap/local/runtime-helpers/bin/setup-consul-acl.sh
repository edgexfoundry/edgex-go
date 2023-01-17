#!/bin/bash
# note: -e flag is not used in this one-shot service
# we don't want to exit out the whole Consul process when ACL bootstrapping failed, just that
# Consul won't have ACL to be used

# setup Consul's ACL via security-bootstrapper's subcommand
"$SNAP"/bin/security-bootstrapper -configDir "$SNAP_DATA"/config/security-bootstrapper/res setupRegistryACL
setupACL_code=$?
if [ "${setupACL_code}" -ne 0 ]; then
  echo "$(date) failed to set up Consul ACL"
fi
