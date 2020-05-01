#!/bin/bash -ex

security-secretstore-setup -confdir "$SNAP_DATA/config/security-secretstore-setup/res" --vaultInterval=10

# if redis pw file doesn't exist, handle initialization & read it
if [ ! -f "$REDIS5_PASSWORD_PATHNAME" ]; then

    # read the pw from secretstore & write to disk
    security-secretstore-read -confdir "$SNAP_DATA/config/security-secretstore-read/res"
fi

