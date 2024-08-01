#!/bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (C) 2024 IOTech Ltd
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

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on Postgres"

# gating on the TokensReadyPort
echo "$(date) Executing waitFor on Postgres with waiting on TokensReadyPort \
  tcp://${STAGEGATE_SECRETSTORESETUP_HOST}:${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}"
/edgex-init/security-bootstrapper --configDir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_SECRETSTORESETUP_HOST}":"${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

# the configurePostgres retrieves the postgres user's credentials from secretstore (i.e. Vault)
echo "$(date) ${STAGEGATE_SECRETSTORESETUP_HOST} tokens ready, bootstrapping postgres..."
/edgex-init/security-bootstrapper --configDir=/edgex-init/bootstrap-postgres/res configurePostgres

postgres_bootstrapping_status=$?
if [ $postgres_bootstrapping_status -ne 0 ]; then
  echo "$(date) failed to bootstrap postgres"
  exit 1
fi


if [ ! -f "${DATABASECONFIG_PATH}"/"${DATABASECONFIG_NAME}" ]; then
  ehco "$(date) Error: initialization script file ${DATABASECONFIG_PATH}/${DATABASECONFIG_NAME} not exists"
  exit 1
fi

# starting postgres
echo "$(date) Starting edgex-postgres ..."
exec /usr/local/bin/docker-entrypoint.sh postgres "$@"
