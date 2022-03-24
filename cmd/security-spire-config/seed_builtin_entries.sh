#!/bin/sh -x
#
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
#

local_agent_svid=$1

echo "local_agent_svid=${local_agent_svid}"
echo "SPIFFE_SERVER_SOCKET=${SPIFFE_SERVER_SOCKET}"
echo "SPIFFE_EDGEX_SVID_BASE=${SPIFFE_EDGEX_SVID_BASE}"

# will replace security-test-client with all core-services later
for service in security-spiffe-token-provider security-test-client; do
    spire-server entry create -socketPath "${SPIFFE_SERVER_SOCKET}" -parentID "${local_agent_svid}" -dns "edgex-${service}" -spiffeID "${SPIFFE_EDGEX_SVID_BASE}/${service}" -selector "docker:label:com.docker.compose.service:${service}"
done
