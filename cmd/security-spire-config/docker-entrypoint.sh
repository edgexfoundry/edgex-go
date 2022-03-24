#!/bin/sh -xe
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

umask 027

export SPIFFE_SERVER_SOCKET
export SPIFFE_EDGEX_SVID_BASE
export SPIFFE_TRUSTDOMAIN

: ${SPIFFE_SERVER_SOCKET:=/tmp/edgex/secrets/spiffe/private/api.sock}
: ${SPIFFE_EDGEX_SVID_BASE:=spiffe://edgexfoundry.org/service}
: ${SPIFFE_TRUSTDOMAIN:=edgexfoundry.org}

: ${SPIFFE_AGENT0_CN:=agent0}
: ${SPIFFE_PARENTID:=spiffe://${SPIFFE_TRUSTDOMAIN}/spire/agent/x509pop/cn/${SPIFFE_AGENT0_CN}}

# Wait for agent CA creation

while test ! -S "${SPIFFE_SERVER_SOCKET}"; do
    echo "Waiting for ${SPIFFE_SERVER_SOCKET}"
    sleep 1
done

# Run config scripts

for script in /usr/local/etc/spiffe-scripts.d/* ; do
    test -x "${script}" && ${script} "${SPIFFE_PARENTID}"
done

exec tail -f /dev/null
