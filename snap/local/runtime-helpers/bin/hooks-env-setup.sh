#!/bin/bash -e

# This script is mainly necessary so from the hooks we can use jq (which is in
# the snap) which requires libs from inside the snap, but hooks don't get the
# same wrapper scripts as apps in snapcraft.yaml do

# add things from the snap's $PATH to here
export PATH="$SNAP/usr/sbin:$SNAP/usr/bin:$SNAP/sbin:$SNAP/bin:$PATH"

# setup LD_LIBRARY_PATH, we need to handle the different architecture paths -
# this snippet is copied out of one of the generated snapcraft wrappers
case $(arch) in
    x86_64)
        MULTI_ARCH_PATH="x86_64-linux-gnu";;
    arm*)
        MULTI_ARCH_PATH="arm-linux-gnueabihf";;
    aarch64)
        MULTI_ARCH_PATH="aarch64-linux-gnu";;
    *)
        echo "architecture $ARCH not supported"
        exit 1
        ;;
esac

# Note - the env var SNAP_LIBRARY_PATH is set by snapd for hooks, so it can be referenced here
export LD_LIBRARY_PATH="$SNAP_LIBRARY_PATH:$LD_LIBRARY_PATH:$SNAP/lib:$SNAP/usr/lib:$SNAP/lib/$MULTI_ARCH_PATH:$SNAP/usr/lib/$MULTI_ARCH_PATH"
