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

# This is customized entrypoint script for proxy-setup.
# In particular, it waits for ready-to-run port and kong to be ready

set -e

# env settings are populated from env files of docker-compose

echo "Script for waiting security bootstrapping on proxy-setup"

# gating on the ready-to-run port
echo "$(date) Executing waitFor for ${PROXY_SETUP_HOST} with waiting on \
  tcp://${STAGEGATE_BOOTSTRAPPER_HOST}:${STAGEGATE_READY_TORUNPORT}"
/edgex-init/security-bootstrapper --confdir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_BOOTSTRAPPER_HOST}":"${STAGEGATE_READY_TORUNPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

echo "$(date) ${PROXY_SETUP_HOST} waits on Kong to be initialized"

kong_inited=0
until [ $kong_inited -eq 1 ]; do
  status=$(/edgex-init/security-bootstrapper --confdir=/edgex-init/res getHttpStatus \
    --url=http://"${API_GATEWAY_HOST}":"${API_GATEWAY_STATUS_PORT}"/status | tail -n 1)
  if [ ${#status} -gt 0 ] && [[ "${status}" != *ERROR* ]]; then
    echo "$(date) ${API_GATEWAY_HOST}:${API_GATEWAY_STATUS_PORT} status code = ${status}"
    if [ "$status" -eq 200 ]; then
      kong_inited=1
    fi
  fi
  if [ $kong_inited -ne 1 ]; then
    echo "$(date) waiting for ${API_GATEWAY_HOST} to be initialized"
    sleep 1
  fi
done

echo "$(date) ${API_GATEWAY_HOST} is ready"

echo "$(date) Starting ${PROXY_SETUP_HOST} ..."
exec /usr/local/bin/entrypoint.sh "$*"
