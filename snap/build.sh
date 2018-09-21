#!/bin/bash -e

# get the directory of this script
# snippet from https://stackoverflow.com/a/246128/10102404
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# build the container image - switch on which architecture we're on
# note that for now the armhf and i386 builds fail, but if we ever support those targets
# this will "just work"
ARCH=$(arch)
if [ $ARCH = "x86_64" ] ; then
    docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.amd64.build .
elif [ $ARCH = "armhf" ] ; then
    docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.armhf.build .
elif [ $ARCH = "aarch64" ] ; then
    docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.arm64.build .
elif [ $ARCH = "i386" ] ; then
    docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.i386.build .
fi

# now run the build
docker run --rm -v $(readlink -f ${SCRIPT_DIR}/..):/build edgex-snap-builder:latest
