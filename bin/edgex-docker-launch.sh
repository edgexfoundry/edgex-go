#!/usr/bin/env sh

# Copyright 2020 Redis Labs Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software distributed under the License
# is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
# or implied. See the License for the specific language governing permissions and limitations under
# the License.

# Usage: bin/edgex-docker-launch.sh [mongo]
#
# By default download the Redis based Docker Compose file and attempt to start EdgeX. If the mongo
# option is used, download the Mongo based Docker Compose file and attemp to start EdgeX.
#
# To override the compose file entirely set the COMPOSE_FILE_PATH environment variable to the full
# pathname of the compose file you want to use.

RELEASE=nightly-build
VERSION=nexus
GITHUB_PATH=https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/releases/"${RELEASE}"/compose-files
PERSIST=${1:-redis}

if [ -z "${COMPOSE_FILE_PATH}" ]; then
    COMPOSE_FILE=docker-compose-${VERSION}-${PERSIST}-no-secty.yml
    
    COMPOSE_FILE_PATH=/tmp/${COMPOSE_FILE}
    mkdir -p "$(dirname "${COMPOSE_FILE_PATH}")"
    echo "Downloading Docker Compose file..."
    curl -s -o "${COMPOSE_FILE_PATH}" ${GITHUB_PATH}/"${COMPOSE_FILE}"
fi

echo "Starting EdgeX..."
docker-compose -f "${COMPOSE_FILE_PATH}" up -d
