#!/usr/bin/dumb-init /bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2019 Intel Corporation
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

VAULT_TLS_PATH=${VAULT_TLS_PATH:-/tmp/edgex/secrets/edgex-vault}

DEFAULT_VAULT_LOCAL_CONFIG='
listener "tcp" { 
              address = "edgex-vault:8200" 
              tls_disable = "0" 
              cluster_address = "edgex-vault:8201" 
              tls_min_version = "tls12" 
              tls_client_ca_file ="'${VAULT_TLS_PATH}'/ca.pem" 
              tls_cert_file ="'${VAULT_TLS_PATH}'/server.crt" 
              tls_key_file = "'${VAULT_TLS_PATH}'/server.key" 
              tls_perfer_server_cipher_suites = "true"
          } 
          backend "consul" { 
              path = "vault/" 
              address = "edgex-core-consul:8500" 
              scheme = "http" 
              redirect_addr = "https://edgex-vault:8200" 
              cluster_addr = "https://edgex-vault:8201" 
          } 
          default_lease_ttl = "168h" 
          max_lease_ttl = "720h"
'

VAULT_LOCAL_CONFIG=${VAULT_LOCAL_CONFIG:-$DEFAULT_VAULT_LOCAL_CONFIG}

echo "VAULT_LOCAL_CONFIG:" ${VAULT_LOCAL_CONFIG}

export VAULT_TLS_PATH VAULT_LOCAL_CONFIG

# Before Vault can start up, the TLS certificates are required.
# By default, these certificates are generated via security-secrets-setup
# application service and the sentinel files are produced when it is done.
# Thus, we check the sentinel file here:

SENTILNEL_FILE=${VAULT_TLS_PATH}/.security-secrets-setup.complete
while test ! -f ${SENTILNEL_FILE}; do
    echo 'waiting for security-secrets-setup done';
    sleep 1;
done;

echo 'Starting edgex-vault...'
exec /usr/local/bin/docker-entrypoint.sh server -log-level=info
