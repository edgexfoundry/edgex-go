#!/bin/sh
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

# This is customized entrypoint script for Redis.
# In particular, it waits for the TokensReady Port being ready to roll

set -e

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on Redis"

# gating on the TokensReadyPort
echo "$(date) Executing waitFor on Redis with waiting on TokensReadyPort \
  tcp://${STAGEGATE_SECRETSTORESETUP_HOST}:${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}"
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_SECRETSTORESETUP_HOST}":"${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

# the configureRedis retrieves the redis default user's credentials from secretstore (i.e. Vault) and 
# generates the redis configuration file with ACL rules in it.
# The redis database server will start with the generated configuration file so that it is 
# started securely.
echo "$(date) ${STAGEGATE_SECRETSTORESETUP_HOST} tokens ready, bootstrapping redis..."
/edgex-init/security-bootstrapper --confdir=/edgex-init/bootstrap-redis/res configureRedis

redis_bootstrapping_status=$?
if [ $redis_bootstrapping_status -ne 0 ]; then
  echo "$(date) failed to bootstrap redis"
  exit 1
fi

# make sure the config file is present before redis server starts up
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri file://"${DATABASECONFIG_PATH}"/"${DATABASECONFIG_NAME}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

# starting redis with config file
echo "$(date) Starting edgex-redis ..."
exec /usr/local/bin/docker-entrypoint.sh redis-server "${DATABASECONFIG_PATH}"/"${DATABASECONFIG_NAME}" &

# wait for the Redis port
echo "$(date) Executing waitFor on database redis with waiting on its own port \
  tcp://${STAGEGATE_DATABASE_HOST}:${STAGEGATE_DATABASE_PORT}"
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_DATABASE_HOST}":"${STAGEGATE_DATABASE_PORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

echo "$(date) redis is bootstrapped and ready"

# Signal that Redis is ready for services blocked waiting on Redis
/edgex-init/security-bootstrapper --confdir=/edgex-init/res listenTcp \
  --port="${STAGEGATE_DATABASE_READYPORT}" --host="${DATABASES_PRIMARY_HOST}"
if [ $? -ne 0 ]; then
  echo "$(date) failed to gating the redis ready port, exits"
fi
