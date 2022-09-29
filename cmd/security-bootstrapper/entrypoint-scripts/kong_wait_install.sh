#!/usr/bin/env bash
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

# This is customized entrypoint script for Kong.
# In particular, it waits for the security-bootstrapper's ReadyToRunPort and Postgres db ready to roll

set -e

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on Kong"

# gating on the ready-to-run port
echo "$(date) Executing waitFor with waiting on tcp://${STAGEGATE_BOOTSTRAPPER_HOST}:${STAGEGATE_READY_TORUNPORT}"
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_BOOTSTRAPPER_HOST}":"${STAGEGATE_READY_TORUNPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

echo "$(date) Kong waits on Postgres to be initialized"
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_KONGDB_HOST}":"${STAGEGATE_KONGDB_READYPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

# KONG_PG_PASSWORD_FILE is env used by Kong, it is for kong-db's password file
echo "$(date) Executing waitFor with waiting on file:${KONG_PG_PASSWORD_FILE}"
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri file://"${KONG_PG_PASSWORD_FILE}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

# double check and make sure the postgres is setup with that password and ready
passwd=$(cat "${KONG_PG_PASSWORD_FILE}")
pg_inited=0
until [ $pg_inited -eq 1 ]; do
  status=$(/edgex-init/security-bootstrapper --confdir=/edgex-init/res pingPgDb \
    --username=kong --dbname=kong --password="${passwd}" | tail -n 1)
  if [ ${#status} -gt 0 ] && [[ "${status}" != *ERROR* ]]; then
    if [ "${status}" = "ready" ]; then
      pg_inited=1
      passwd=""
    fi
  fi
  if [ $pg_inited -ne 1 ]; then
    echo "$(date) waiting for ${STAGEGATE_KONGDB_HOST} to be initialized"
    sleep 1
  fi
done

echo "$(date) Check point: postgres db is ready for kong"

# in kong's docker, we use KONG_PG_PASSWORD_FILE instead of KONG_PG_PASSWORD for better security
export KONG_PG_PASSWORD_FILE

# remove env KONG_PG_PASSWORD: only use KONG_PG_PASSWORD_FILE
unset KONG_PG_PASSWORD

set +e
/docker-entrypoint.sh kong migrations bootstrap
/docker-entrypoint.sh kong migrations up
/docker-entrypoint.sh kong migrations finish
code=$?
if [ $code -ne 0 ]; then
  echo "$(date) failed to kong migrations, returned code = " $code
  exit $code
fi
set -e
echo "$(date) Configuring Kong Admin API..."

# Running "kong config db_import" will return a non-successful error code even though it 
# has successfully run. This is why we added "|| true" to the end of this call.
/docker-entrypoint.sh kong config db_import /usr/local/kong/kong.yml || true

echo "$(date) Starting kong ..."
exec /docker-entrypoint.sh kong docker-start
