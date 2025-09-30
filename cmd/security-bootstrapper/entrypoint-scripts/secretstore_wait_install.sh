#!/usr/bin/dumb-init /bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2021 Intel Corporation
#  Copyright (c) 2024 IOTech Ltd
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

# This is customized entrypoint script for secret store.
# In particular, it waits for the BootstrapPort ready to roll

set -e

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on Secret Store"

DEFAULT_BAO_LOCAL_CONFIG='
listener "tcp" { 
              address = "edgex-secret-store:8200"
              tls_disable = "1" 
              cluster_address = "edgex-secret-store:8201"
          } 
          backend "file" {
              path = "/openbao/file"
          } 
          default_lease_ttl = "168h" 
          max_lease_ttl = "720h"
'

BAO_LOCAL_CONFIG=${BAO_LOCAL_CONFIG:-$DEFAULT_BAO_LOCAL_CONFIG}

DEFAULT_BAO_LOG_LEVEL='info'
BAO_LOG_LEVEL=${BAO_LOG_LEVEL:-$DEFAULT_BAO_LOG_LEVEL}

export BAO_LOCAL_CONFIG

echo "$(date) BAO_LOCAL_CONFIG: ${BAO_LOCAL_CONFIG}"

if [ "$1" = 'server' ]; then
  echo "$(date) Executing waitFor on secret store $* with \
    tcp://${STAGEGATE_BOOTSTRAPPER_HOST}:${STAGEGATE_BOOTSTRAPPER_STARTPORT}"
  /edgex-init/security-bootstrapper --configDir=/edgex-init/res waitFor \
    -uri tcp://"${STAGEGATE_BOOTSTRAPPER_HOST}":"${STAGEGATE_BOOTSTRAPPER_STARTPORT}" \
    -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

  echo "$(date) Starting edgex-secret-store..."
  exec /usr/local/bin/docker-entrypoint.sh server -log-level=${BAO_LOG_LEVEL}
fi
