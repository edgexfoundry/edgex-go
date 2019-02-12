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

    # figure out what kind of job this is using $JOB_NAME and simplify that 
    # into $JOB_TYPE
    JOB_TYPE="build"
    if [[ "$JOB_NAME" =~ edgex-go-snap-.*-stage-snap.* ]]; then
        JOB_TYPE="stage"
    elif [[ "$JOB_NAME" =~ edgex-go-snap-release-snap ]]; then
        JOB_TYPE="release"
    fi
fi

# build the container image - providing the relevant architecture we're on
# to determine which snap arch to download in the docker container
case $(arch) in 
    x86_64)
        arch="amd64";;
    aarch64)
        arch="arm64";;
    arm*)
        arch="armhf";;
esac
docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.build --build-arg ARCH="$arch" $GIT_ROOT

# delete the login file we copied to the git root so it doesn't persist around
rm $GIT_ROOT/edgex-snap-store-login

# now run the build with the environment variables 
docker run --rm -e "JOB_TYPE=$JOB_TYPE" -e "SNAP_REVISION=$SNAP_REVISION" -e "SNAP_CHANNEL=$SNAP_CHANNEL" edgex-snap-builder:latest

# note that we don't need to delete the docker images here, that's done for us by jenkins in the 
# edgex-provide-docker-cleanup macro defined for all the snap jobs
