#!/bin/bash -e

# This script is called as a post-stop-command when
# security-secretstore-setup oneshot service stops.
#
# In addition, it may be called by connect-plug-edgex-secretstore-token hook
# before or after security-secretstore-setup. For details, refer to the
# respective hook script.

# The caller is either post-stop-command from snapcraft.yaml (default)
# or the connect-plug-edgex-secretstore-token hook.
# It is only for logging purposes
caller=${1:-"post-stop-command"}
logger "edgex-secretstore:post-setup: started by $caller"

# create the directory which consumers bind-mount into
mkdir -p $SNAP_DATA/mount/secrets

# Each directory corresponds to an external device/app service that is connected
# to the edgex-secretstore-token plug
for fpath in $SNAP_DATA/mount/secrets/*; do
    # verify that this is a directory
    [ -d "$fpath" ] || continue

    # service name must be the same as the directory name
    fname=$(basename "$fpath")
    # path to where the token for this service is generated
    TOKEN=$SNAP_DATA/secrets/$fname/secrets-token.json
    # bind mount target directory path for the copy incl. trailing slash
    SECRETS_MOUNT_DIR=$SNAP_DATA/mount/secrets/$fname/
    
    if [ -f "$TOKEN" ]; then
        logger "edgex-secretstore:post-setup: copying $TOKEN to $SECRETS_MOUNT_DIR"
        cp -vr $TOKEN $SECRETS_MOUNT_DIR
    else
        # This is logged for interfaces that are auto-connected before the 
        # security-secretstore-setup runs for the first time, because tokens
        # are not yet available.
        #
        # It is an error if security-secretstore-setup has already run but
        # the expected token wasn't generated due to an internal error.
        #
        # It is also an error if the consumer is connecting to receive a token 
        # that hasn't been generated per configuration.
        #
        # Regardless of the error cases, this should not be raised to an error
        # and exit with non-zero code because it prevents the installation of 
        # this snap (auto-connection error) for cases that are beyond the
        # control of this snap.
        logger "edgex-secretstore:post-setup: could not find token for $fname"
    fi
done
