#!/bin/bash -e

# Required by click.
export LC_ALL=C.UTF-8
export SNAPCRAFT_SETUP_CORE=1

# this tells snapcraft to include a manifest file in the snap
# detailing which packages were used to build the snap
export SNAPCRAFT_BUILD_INFO=1

# if snapcraft ever encounters any bugs, we should force it to 
# auto-report silently rather than attempt to ask for permission
# to send a report
export SNAPCRAFT_ENABLE_SILENT_REPORT=1


# clean the environment and build the snap
build_snap()
{
    pushd /build > /dev/null
    snapcraft clean
    snapcraft
    popd > /dev/null
}

# login to the snap store using the provided login macaroon file
snapcraft_login()
{    
    snapcraft login --with /build/edgex-snap-store-login
}

# release a locally build snap to the store
release_local_snap() 
{
    pushd /build > /dev/null
    snapcraft_login
    # push the snap up to the store and get the revision of the snap
    REVISION=$(snapcraft push edgexfoundry*.snap | awk '/Revision/ {print $2}')
    # now release it on the provided revision and snap channel
    snapcraft release edgexfoundry $REVISION $SNAP_CHANNEL 
    # also update the meta-data automatically
    snapcraft push-metadata edgexfoundry*.snap --force
    popd > /dev/null
}

# release a snap revision already in the store
release_store_snap()
{
    snapcraft_login
    snapcraft release edgexfoundry $SNAP_REVISION $SNAP_CHANNEL 
}

case "$JOB_TYPE" in 
    "stage")
        # stage jobs build the snap locally and release it
        build_snap
        release_local_snap
    ;;
    "release")
        # release jobs will promote an already built snap revision
        # in the store to a channel
        release_store_snap
    ;;
    *)
        # do normal build and nothing else
        build_snap
    ;;
esac
