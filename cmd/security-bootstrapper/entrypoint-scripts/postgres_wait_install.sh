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

# This is customized entrypoint script for Postgres.
# In particular, it waits for Vault to be ready and dynamically generated password seeded into db

set -e

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on Postgres"

# Postgres is waiting for BOOTSTRAP_PORT
echo "$(date) Executing waitFor on Postgres with waiting on \
  tcp://${STAGEGATE_BOOTSTRAPPER_HOST}:${STAGEGATE_BOOTSTRAPPER_STARTPORT}"
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_BOOTSTRAPPER_HOST}":"${STAGEGATE_BOOTSTRAPPER_STARTPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

echo "$(date) Postgres waits on Vault to be initialized"

vault_inited=0
until [ $vault_inited -eq 1 ]; do
  status=$(/edgex-init/security-bootstrapper --confdir=/edgex-init/res getHttpStatus \
    --url=http://"${SECRETSTORE_HOST}":"${SECRETSTORE_PORT}"/v1/sys/health | tail -n 1)
  if [ ${#status} -gt 0 ] && [[ "${status}" != *ERROR* ]]; then
    echo "$(date) ${SECRETSTORE_HOST}:${SECRETSTORE_PORT} status code = ${status}"
    if [ "$status" -eq 200 ]; then
      vault_inited=1
    fi
  fi
  if [ $vault_inited -ne 1 ]; then
    echo "$(date) waiting for ${SECRETSTORE_HOST} to be initialized"
    sleep 1
  fi
done

echo "$(date) ${SECRETSTORE_HOST} is ready"

# POSTGRES_PASSWORD_FILE env is used by Postgres and it is for the db password file
# if password already in then re-use
if [ -n "${POSTGRES_PASSWORD_FILE}" ] && [ -f "${POSTGRES_PASSWORD_FILE}" ]; then
  echo "$(date) previous file already exists, skipping creation"
else
  # create password file for postgres to be used in the compose file
  mkdir -p "$(dirname "${POSTGRES_PASSWORD_FILE}")"
  out=$(/edgex-init/security-bootstrapper --confdir=/edgex-init/res genPassword | tail -n 1)
  if [ ${#out} -gt 0 ] && [[ "${out}" != *ERROR* ]]; then
    echo "${out}" > "${POSTGRES_PASSWORD_FILE}"
  fi
fi

chmod 444 "${POSTGRES_PASSWORD_FILE}"
export POSTGRES_PASSWORD_FILE

echo "$(date) Starting kong-db..."
exec /usr/local/bin/docker-entrypoint.sh postgres &

# check that the postgres is initialized
passwd=$(cat "${POSTGRES_PASSWORD_FILE}")
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

echo "$(date) ${STAGEGATE_KONGDB_HOST} is initialized"

# Signal that Postgres is ready for services blocked waiting on Postgres
exec su-exec postgres /edgex-init/security-bootstrapper --confdir=/edgex-init/res listenTcp \
  --port="${STAGEGATE_KONGDB_READYPORT}" --host="${STAGEGATE_KONGDB_HOST}"
if [ $? -ne 0 ]; then
  echo "$(date) failed to gating the postgres ready port, exits"
fi
