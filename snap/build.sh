#!/bin/bash

# get the directory of this script
# snippet from https://stackoverflow.com/a/246128/10102404
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# get the git root, which is one directory up from this script
GIT_ROOT=$(readlink -f ${SCRIPT_DIR}/..)

# if we are running inside a jenkins instance then copy the login file 
# and check if this is a release job
if [ ! -z "$JENKINS_URL" ]; then
    if [ -f $HOME/EdgeX ]; then
       cp $HOME/EdgeX $GIT_ROOT/edgex-snap-store-login
    else
        echo "I seem to be running on Jenkins, but there's not a snap store login file..." 
    fi

    # check if this is a release job or not, if it is set the corresponding env var
    if [[ "$JOB_NAME" =~ edgex-go-snap-.*-stage-snap.* ]]; then
        IS_RELEASE_JOB="YES"
    else
        IS_RELEASE_JOB="NO"
    fi
fi

# build the container image - switch on which architecture we're on
# note that for now the armhf and i386 builds fail, but if we ever support those targets
# this will "just work"
ARCH=$(arch)
if [ $ARCH = "x86_64" ] ; then
    docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.amd64.build $GIT_ROOT
elif [ $ARCH = "armhf" ] ; then
    docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.armhf.build $GIT_ROOT
elif [ $ARCH = "aarch64" ] ; then
    docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.arm64.build $GIT_ROOT
elif [ $ARCH = "i386" ] ; then
    docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.i386.build $GIT_ROOT
fi

# delete the login file we copied to the git root so it doesn't persist around
rm $GIT_ROOT/edgex-snap-store-login

# now run the build with the environment variables 
docker run --rm -e "IS_RELEASE_JOB=$IS_RELEASE_JOB" -e "RELEASE=$RELEASE" -e "SNAP_CHANNEL=$SNAP_CHANNEL" edgex-snap-builder:latest

# note that we don't need to delete the docker images here, that's done for us by jenkins in the 
# edgex-provide-docker-cleanup macro defined for all the snap jobs
