#!/usr/bin/dumb-init /bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2020 Intel Corporation
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#
#  SPDX-License-Identifier: Apache-2.0'
#  ----------------------------------------------------------------------------------

set -e

if [ -n "${SECRETSTORE_SETUP_DONE_FLAG}" ] && [ -f "${SECRETSTORE_SETUP_DONE_FLAG}" ]; then
  echo "Clearing secretstore-setup completion flag"
  rm -f "${SECRETSTORE_SETUP_DONE_FLAG}"
fi

echo "Starting vault-worker..."

echo "Initializing secret store..."
/security-secretstore-setup --vaultInterval=10

# write a sentinel file when we're done because consul is not
# secure and we don't trust it it access to the EdgeX secret store
if [ -n "${SECRETSTORE_SETUP_DONE_FLAG}" ]; then
    # default User and Group in case never set
    if [ -z "${EDGEX_USER}" ]; then
      EDGEX_USER="2002"
      EDGEX_GROUP="2001"
    fi

    echo "Changing ownership of secrets to ${EDGEX_USER}:${EDGEX_GROUP}"
    chown -Rh ${EDGEX_USER}:${EDGEX_GROUP} /tmp/edgex/secrets

    echo "Signaling secretstore-setup completion"
    mkdir -p $(dirname "${SECRETSTORE_SETUP_DONE_FLAG}") && \
      touch "${SECRETSTORE_SETUP_DONE_FLAG}"
fi

echo "Waiting for termination signal"
exec tail -f /dev/null
