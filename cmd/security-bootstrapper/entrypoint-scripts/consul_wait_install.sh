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

# This is customized entrypoint script for Consul and run on the consul's container
# In particular, it waits for Vault to be ready to roll

set -e

# function to check on Vault for readiness
vault_ready()
{
  vault_host=$1
  vault_port=$2
  resp_code=$(curl --write-out '%{http_code}' --silent --output /dev/null "${vault_host}":"${vault_port}"/v1/sys/health)
  if [ "$resp_code" -eq 200 ] ; then
    echo 1
  else
    echo 0
  fi
}

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on Consul"

echo "$(date) Consul waits on Vault to be initialized"
# check the http status code from Vault using SECRETSTORE_HOST and SECRETSTORE_PORT as input to the function call
vault_inited=$(vault_ready "${SECRETSTORE_HOST}" "${SECRETSTORE_PORT}")
until [ "$vault_inited" -eq 1 ]; do
    echo "$(date) waiting for Vault ${SECRETSTORE_HOST}:${SECRETSTORE_PORT} to be initialized";
    sleep 1;
    vault_inited=$(vault_ready "${SECRETSTORE_HOST}" "${SECRETSTORE_PORT}")
done

# only in json format according to Consul's documentation
DEFAULT_CONSUL_LOCAL_CONFIG='
{
    "enable_local_script_checks": true,
    "disable_update_check": true
}
'

# set the default value to environment var if not present
CONSUL_LOCAL_CONFIG=${CONSUL_LOCAL_CONFIG:-$DEFAULT_CONSUL_LOCAL_CONFIG}

export CONSUL_LOCAL_CONFIG

echo "$(date) CONSUL_LOCAL_CONFIG: ${CONSUL_LOCAL_CONFIG}"

echo "$(date) Starting edgex-consul..."
exec docker-entrypoint.sh agent -ui -bootstrap -server -client 0.0.0.0 &

# wait for the consul port
echo "$(date) Executing dockerize on Consul with waiting on its own port \
  tcp://${STAGEGATE_REGISTRY_HOST}:${STAGEGATE_REGISTRY_PORT}"
/edgex-init/dockerize -wait tcp://"${STAGEGATE_REGISTRY_HOST}":"${STAGEGATE_REGISTRY_PORT}" \
  -timeout "${SECTY_BOOTSTRAP_GATING_TIMEOUT_DURATION}"

# Signal that Consul is ready for services blocked waiting on Consul
/edgex-init/security-bootstrapper --confdir=/edgex-init/res listenTcp \
  --port="${STAGEGATE_REGISTRY_READYPORT}" --host="${STAGEGATE_REGISTRY_HOST}"
if [ $? -ne 0 ]; then
    echo "$(date) failed to gating the consul ready port, exits"
fi
