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

# only in json format according to Consul's documentation
DEFAULT_CONSUL_LOCAL_CONFIG='
{
    "enable_local_script_checks": true,
    "disable_update_check": true,
    "ports": {
      "dns": -1
    }
}
'

# set the default value to environment var if not present
CONSUL_LOCAL_CONFIG=${CONSUL_LOCAL_CONFIG:-$DEFAULT_CONSUL_LOCAL_CONFIG}

export CONSUL_LOCAL_CONFIG

echo "$(date) CONSUL_LOCAL_CONFIG: ${CONSUL_LOCAL_CONFIG}"

echo "$(date) Starting edgex-core-consul with ACL enabled ..."
docker-entrypoint.sh agent \
  -ui \
  -bootstrap \
  -server \
  -config-file=/edgex-init/consul-bootstrapper/config_consul_acl.json \
  -client 0.0.0.0 &
# wait for the secretstore tokens ready as we need the token for bootstrapping
echo "$(date) Executing waitFor on Consul with waiting on TokensReadyPort \
  tcp://${STAGEGATE_SECRETSTORESETUP_HOST}:${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}"
/edgex-init/security-bootstrapper --configDir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_SECRETSTORESETUP_HOST}":"${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

# we don't want to exit out the whole Consul process when ACL bootstrapping failed, just that
# Consul won't have ACL to be used
set +e
# call setupRegistryACL bootstrapping command, containing both ACL bootstrapping and re-configure consul access steps
/edgex-init/security-bootstrapper --configDir=/edgex-init/res setupRegistryACL
setupACL_code=$?
if [ "${setupACL_code}" -ne 0 ]; then
  echo "$(date) failed to set up Consul ACL"
fi

# we need to grant the permission for proxy setup to read consul's token path so as to retrieve consul's token from it
echo "$(date) Changing ownership of consul token path to ${EDGEX_USER}:${EDGEX_GROUP}"
chown -Rh "${EDGEX_USER}":"${EDGEX_GROUP}" "${STAGEGATE_REGISTRY_ACL_MANAGEMENTTOKENPATH}"
set -e
# no need to wait for Consul's port since it is in ready state after all ACL stuff

# Signal that Consul is ready for services blocked waiting on Consul
exec su-exec consul /edgex-init/security-bootstrapper --configDir=/edgex-init/res listenTcp \
  --port="${STAGEGATE_REGISTRY_READYPORT}" --host="${STAGEGATE_REGISTRY_HOST}"
if [ $? -ne 0 ]; then
    echo "$(date) failed to gating the consul ready port, exits"
fi
