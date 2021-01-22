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
echo "$(date) Executing dockerize on Redis with waiting on TokensReadyPort \
  tcp://${STAGEGATE_SECRETSTORESETUP_HOST}:${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}"
/edgex-init/dockerize -wait tcp://"${STAGEGATE_SECRETSTORESETUP_HOST}":"${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}" \
  -timeout "${SECTY_BOOTSTRAP_GATING_TIMEOUT_DURATION}"

# the bootstrap-redis needs the connection from Redis db to set it up.
# Hence, here bootstrap-redis runs in background and then after bootstrap-redis starts,
# the Redis db starts in background.
echo "$(date) ${STAGEGATE_SECRETSTORESETUP_HOST} tokens ready, bootstrapping redis..."
/edgex-init/bootstrap-redis/security-bootstrap-redis --confdir=/edgex-init/bootstrap-redis/res &
redis_bootstrapper_pid=$!

# give some time for bootstrap-redis to start up
sleep 1

echo "$(date) Starting edgex-redis..."
exec /usr/local/bin/docker-entrypoint.sh redis-server &

# wait for bootstrap-redis to finish before signal the redis is ready
wait $redis_bootstrapper_pid
redis_bootstrapping_status=$?
if [ $redis_bootstrapping_status -eq 0 ]; then
  echo "$(date) redis is bootstrapped and ready"
else
  echo "$(date) failed to bootstrap redis"
fi

# Signal that Redis is ready for services blocked waiting on Redis
/edgex-init/security-bootstrapper --confdir=/edgex-init/res listenTcp \
  --port="${STAGEGATE_DATABASE_READYPORT}" --host="${DATABASES_PRIMARY_HOST}"
if [ $? -ne 0 ]; then
  echo "$(date) failed to gating the redis ready port, exits"
fi
