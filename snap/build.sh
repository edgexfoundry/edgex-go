#!/bin/bash -e

# get the directory of this script
# snippet from https://stackoverflow.com/a/246128/10102404
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# build the container image
docker build -t edgex-snap-builder:latest -f ${SCRIPT_DIR}/Dockerfile.build .

# now run the build
docker run --rm -v $(readlink -f ${SCRIPT_DIR}/..):/build edgex-snap-builder:latest
