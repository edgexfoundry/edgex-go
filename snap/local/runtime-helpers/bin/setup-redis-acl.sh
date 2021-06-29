#!/bin/bash

DBFILE="$DATABASECONFIG_PATH/$DATABASECONFIG_NAME"

logger "setup-redis-acl: redis config file:: $DBFILE"

# setup Consul's ACL via security-bootstrapper's subcommand
"$SNAP"/bin/security-bootstrapper -confdir "$SNAP_DATA"/config/security-bootstrapper/res-bootstrap-redis configureRedis
setupACL_code="$?"
if [ "${setupACL_code}" -ne 0 ]; then
  logger "$(date) failed to set up Redis ACL"
fi

# The redis configuration file contains a path to the ACL
# file found in the same directory. This path is generated
# by security-bootstrapper using $SNAP_DATA, and thus ends
# up with a hard-coded revision which will cause a refresh
# to fail. This sed statement replaces the revision in the
# path with the string 'current'.
if [ -f "$DBFILE" ]; then
    logger "setup-redis-acl: updating ACL path with 'current' symlink"
    sed -i -e "s@foundry\/.*\/redis@foundry\/current\/redis@" "$DBFILE"
fi

