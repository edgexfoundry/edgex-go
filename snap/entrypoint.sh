#!/bin/bash -e

# Required by click.
export LC_ALL=C.UTF-8
export SNAPCRAFT_SETUP_CORE=1

# build the snap
cd /build
snapcraft clean
snapcraft

# only on release jobs release the snap
if [ "$IS_RELEASE_JOB" = "YES" ]; then
    # login using the provided login
    snapcraft login --with /build/edgex-snap-store-login
    # push the snap up to the store and get the revision of the snap
    REVISION=$(snapcraft push edgexfoundry*.snap | awk '/Revision/ {print $2}')
    # now release it on the provided revision and snap channel
    snapcraft release edgexfoundry $REVISION $SNAP_CHANNEL 
fi
