#!/usr/bin/dumb-init /bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2022 Intel Corporation
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
#  SPDX-License-Identifier: Apache-2.0
#  ----------------------------------------------------------------------------------

set -e

# EDGEX_SECRETS_ROOT should default to /tmp/edgex/secrets
# unless changed in configuration.yaml or overridden by environment variable
EDGEX_SECRETS_ROOT=`yq -r .TokenFileProvider.OutputDir /res-file-token-provider/configuration.yaml`
if [ ! -z "${TOKENFILEPROVIDER_OUTPUTDIR}" ]; then
  EDGEX_SECRETS_ROOT="${TOKENFILEPROVIDER_OUTPUTDIR}"
fi

# create token dir, and assign perms
mkdir -p /vault/config/assets
chown -Rh 100:1000 /vault/

echo "Initializing secret store..."
/security-secretstore-setup --vaultInterval=10

# default User and Group in case never set
if [ -z "${EDGEX_USER}" ]; then
  EDGEX_USER="2002"
  EDGEX_GROUP="2001"
fi

# /tmp/edgex/secrets need to be shared with all other services that need secrets and
# thus change the ownership to EDGEX_USER:EDGEX_GROUP
echo "$(date) Changing ownership of ${EDGEX_SECRETS_ROOT} to ${EDGEX_USER}:${EDGEX_GROUP}"
chown -Rh ${EDGEX_USER}:${EDGEX_GROUP} "${EDGEX_SECRETS_ROOT}"

echo "$(date) Changing ownership of ${EDGEX_SECRETS_ROOT}/rules-engine to 1001:1001 to accommodate ekuiper"
chown -Rh 1001:1001 "${EDGEX_SECRETS_ROOT}/rules-engine"

# Signal tokens ready port for other services waiting on
exec su-exec ${EDGEX_USER} /edgex-init/security-bootstrapper --configDir=/edgex-init/res listenTcp \
  --port="${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}" --host="${STAGEGATE_SECRETSTORESETUP_HOST}"
if [ $? -ne 0 ]; then
  echo "$(date) failed to gating the tokens ready port"
fi
