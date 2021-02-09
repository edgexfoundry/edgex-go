#!/usr/bin/dumb-init /bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2021 Intel Corporation
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

# This is customized entrypoint script for Vault.
# In particular, it waits for the BootstrapPort ready to roll

set -e

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on Vault"

DEFAULT_VAULT_LOCAL_CONFIG='
listener "tcp" { 
              address = "edgex-vault:8200" 
              tls_disable = "1" 
              cluster_address = "edgex-vault:8201" 
          } 
          backend "file" {
              path = "/vault/file"
          } 
          default_lease_ttl = "168h" 
          max_lease_ttl = "720h"
'

VAULT_LOCAL_CONFIG=${VAULT_LOCAL_CONFIG:-$DEFAULT_VAULT_LOCAL_CONFIG}

export VAULT_LOCAL_CONFIG

echo "$(date) VAULT_LOCAL_CONFIG: ${VAULT_LOCAL_CONFIG}"

if [ "$1" = 'server' ]; then
  echo "$(date) Executing waitFor on vault $* with \
    tcp://${STAGEGATE_BOOTSTRAPPER_HOST}:${STAGEGATE_BOOTSTRAPPER_STARTPORT}"
  /edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
    -uri tcp://"${STAGEGATE_BOOTSTRAPPER_HOST}":"${STAGEGATE_BOOTSTRAPPER_STARTPORT}" \
    -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

  echo "$(date) Starting edgex-vault..."
  exec /usr/local/bin/docker-entrypoint.sh server -log-level=info
fi
