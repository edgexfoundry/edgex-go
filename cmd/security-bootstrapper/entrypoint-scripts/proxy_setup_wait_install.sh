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

echo "$(date) Starting ${PROXY_SETUP_HOST} ..."
exec /usr/local/bin/entrypoint.sh "$*"
