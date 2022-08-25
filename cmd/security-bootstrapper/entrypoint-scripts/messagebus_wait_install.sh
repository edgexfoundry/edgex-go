#!/bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2022 Intel Corporation
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

# This is customized entrypoint script for message bus
# In particular, it waits for the TokensReady Port being ready to roll

set -e

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on message bus"

# gating on the TokensReadyPort
echo "$(date) Executing waitFor on message bus with waiting on TokensReadyPort \
  tcp://${STAGEGATE_SECRETSTORESETUP_HOST}:${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}"
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_SECRETSTORESETUP_HOST}":"${STAGEGATE_SECRETSTORESETUP_TOKENS_READYPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

# the subcommand snippet retrieves the mqtt's credentials from secretstore (i.e. Vault) and
# generates the configuration files.
# The message broker will start with the generated configuration file so that it is
# started securely.
echo "$(date) ${STAGEGATE_SECRETSTORESETUP_HOST} tokens ready, bootstrapping ${MESSAGE_BUS_TYPE}..."

if [ "$MESSAGE_BUS_TYPE" == "mqtt-bus" ]; then 
  /edgex-init/security-bootstrapper --confdir=/edgex-init/bootstrap-mqtt/res setupMessageBusCreds
  msgbus_bootstrapping_status=$?
  if [ $msgbus_bootstrapping_status -ne 0 ]; then
    echo "$(date) failed to bootstrap ${MESSAGE_BUS_TYPE}"
    exit 1
  fi

  # starting mqtt broker with the pre-config'ed file
  echo "$(date) Starting edgex-${MESSAGE_BUS_TYPE} ..."
  exec /docker-entrypoint.sh /usr/sbin/mosquitto -c /mosquitto/config/mosquitto.conf
else
  echo "Bootstrapping ${MESSAGE_BUS_TYPE} not supported"
  exit 1
fi