#!/bin/sh -x
#
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2022-2023 Intel Corporation
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

echo "EDGEX_SPIFFE_CUSTOM_SERVICES=${EDGEX_SPIFFE_CUSTOM_SERVICES}"

SPIFFE_SERVICES="security-spiffe-token-provider support-notifications support-scheduler \
                 device-bacnet device-camera device-grove device-modbus device-mqtt device-rest device-snmp \
                 device-virtual device-rfid-llrp device-coap device-gpio \
                 app-http-export app-mqtt-export app-sample app-rfid-llrp-inventory \
                 app-external-mqtt-trigger app-metrics-influxdb"

SEED_SERVICES="${SPIFFE_SERVICES} ${EDGEX_SPIFFE_CUSTOM_SERVICES}"

# add pre-authorized services into spire server entry
for dockerservice in $SEED_SERVICES; do
    # Temporary workaround because service name in dockerfile is not consistent with service key.
    # TAF scripts depend on legacy docker-compose service name. Fix in EdgeX 3.0.
    service=`echo -n ${dockerservice} | sed -e 's/app-service-/app-/'`
    spire-server entry create -socketPath "${SPIFFE_SERVER_SOCKET}" -parentID "${local_agent_svid}" -dns "edgex-${service}" -spiffeID "${SPIFFE_EDGEX_SVID_BASE}/${service}" -selector "docker:label:com.docker.compose.service:${dockerservice}"
done

# Always exit successfully even if couldn't (re-)create server entries.
exit 0
