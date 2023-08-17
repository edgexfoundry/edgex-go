#!/bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2021-2023 Intel Corporation
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

set -e

# env settings are populated from env files of docker-compose

echo "Awaiting ReadyToRun signal prior to starting running proxy-setup tasks"

# gating on the ready-to-run port
echo "$(date) Executing waitFor for ${PROXY_SETUP_HOST} with waiting on \
  tcp://${STAGEGATE_BOOTSTRAPPER_HOST}:${STAGEGATE_READY_TORUNPORT}"
/edgex-init/security-bootstrapper --configDir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_BOOTSTRAPPER_HOST}":"${STAGEGATE_READY_TORUNPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"

# The entrypoint script will write the proxy configuration files
echo "$(date) Starting ${PROXY_SETUP_HOST} ..."
/usr/local/bin/entrypoint.sh "$*"

# default User and Group in case never set
if [ -z "${EDGEX_USER}" ]; then
  EDGEX_USER="2002"
  EDGEX_GROUP="2001"
fi

# Signal the reverse proxy that configuration files are set up
echo "$(date) Signalling reverse proxy to start (loop forever) ..."
exec su-exec ${EDGEX_USER} /edgex-init/security-bootstrapper --configDir=/edgex-init/res listenTcp \
  --port="${STAGEGATE_PROXYSETUP_READYPORT}" --host="${PROXY_SETUP_HOST}"
if [ $? -ne 0 ]; then
  echo "$(date) failed to signal the proxy setup ready port"
fi
