#!/bin/bash
#  ----------------------------------------------------------------------------------
#  Copyright (C) 2024-2025 IOTech Ltd
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

# execute security-bootstrapper scripts only with root user
if [ "$(id -u)" = '0' ]; then
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
  find "${DATABASECONFIG_PATH}" \! -user postgres -exec chown postgres '{}' +
  chmod 700 "${DATABASECONFIG_PATH}"

  if [ ! -f "/run/secrets/postgres_password" ]; then
    ehco "$(date) Error: password file /run/secrets/postgres_password not exists"
    exit 1
  fi
  find "/run/secrets" \! -user postgres -exec chown postgres '{}' +
  chmod 700 "/run/secrets"
fi

# customizing of Postgres startup process by including the docker-entrypoint script
source /usr/local/bin/docker-entrypoint.sh

docker_setup_env
docker_create_db_directories

if [ "$(id -u)" = '0' ]; then
	# restart script as postgres user
	exec gosu postgres "$BASH_SOURCE" "$@"
fi

export POSTGRES_PASSWORD_FILE=/run/secrets/postgres_password
PASSWORD=$(<"$POSTGRES_PASSWORD_FILE")
if [ -z "$PASSWORD" ]; then
  echo "$(date) Error: no superuser password define in the /run/secrets/postgres_password file"
  exit 1
fi

# Export POSTGRES_PASSWORD to satisfy the entrypoint script
export POSTGRES_PASSWORD="$PASSWORD"


# run additional initialize db scripts not located in /docker-entrypoint-initdb.d dir if database is initialized for the first time
if [ -z "$DATABASE_ALREADY_EXISTS" ]; then
	docker_verify_minimum_env
	docker_init_database_dir
	pg_setup_hba_conf

	docker_temp_server_start "$@" -c max_locks_per_transaction=256
	docker_setup_db
	docker_process_init_files /docker-entrypoint-initdb.d/*
else
	docker_temp_server_start "$@"

	# Update the superuser password with the value of POSTGRES_PASSWORD
	docker_process_sql <<<"ALTER USER postgres WITH PASSWORD '${POSTGRES_PASSWORD}';"
fi

docker_process_init_files ${DATABASECONFIG_PATH}/*
docker_temp_server_stop

# Remove the POSTGRES_PASSWORD_FILE
rm -f "$POSTGRES_PASSWORD_FILE"

# Check if the file has been removed successfully
if [ -e "$POSTGRES_PASSWORD_FILE" ]; then
    echo "$(date) Failed to remove the POSTGRES_PASSWORD_FILE in: $POSTGRES_PASSWORD_FILE"
fi

# starting postgres
echo "$(date) Starting edgex-postgres ..."
exec postgres "$@"
