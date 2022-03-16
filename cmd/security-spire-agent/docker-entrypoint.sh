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

# Set default env vars if unassigned

: ${SPIFFE_SERVER_SOCKET:=/tmp/edgex/secrets/spiffe/private/api.sock}
: ${SPIFFE_ENDPOINTSOCKET:=/tmp/edgex/secrets/spiffe/public/api.sock}
: ${SPIFFE_TRUSTBUNDLE_PATH:=/tmp/edgex/secrets/spiffe/trust/bundle}
: ${SPIFFE_TRUSTDOMAIN:=edgexfoundry.org}
: ${SPIFFE_SERVER_HOST:=edgex-security-spire-server}
: ${SPIFFE_SERVER_PORT:=59840}
: ${SPIFFE_AGENT0_CN:=agent0}


for dir in `dirname "${SPIFFE_SERVER_SOCKET}"` \
           `dirname "${SPIFFE_TRUSTBUNDLE_PATH}"` \
           `dirname "${SPIFFE_ENDPOINTSOCKET}"` ; do
    test -d "$dir" || mkdir -p "$dir"
done

# Wait for agent CA creation

while test ! -f "/srv/spiffe/ca/public/agent-ca.crt"; do
    echo "Waiting for /srv/spiffe/ca/public/agent-ca.crt"
    sleep 1
done

# Create agent certificate

if test ! -f "/srv/spiffe/agent/agent.crt"; then
    openssl ecparam -genkey -name secp521r1 -noout -out "/srv/spiffe/agent/agent.key"
    SAN="" openssl req -subj "/CN=${SPIFFE_AGENT0_CN}" -config "/usr/local/etc/openssl.conf" -key "/srv/spiffe/agent/agent.key" -sha512 -new -out "/run/agent.req.$$"
    SAN="" openssl x509 -sha512 -extfile /usr/local/etc/openssl.conf -extensions agent_ext -CA "/srv/spiffe/ca/public/agent-ca.crt" -CAkey "/srv/spiffe/ca/private/agent-ca.key" -CAcreateserial -req -in "/run/agent.req.$$" -days 3650 -out "/srv/spiffe/agent/agent.crt"
    rm -f "/run/agent.req.$$"
fi

# Export trust bundle

spire-server bundle show -socketPath "${SPIFFE_SERVER_SOCKET}" > "${SPIFFE_TRUSTBUNDLE_PATH}"
if [ $? -eq 0 ]; then
    echo INFO: Exported trust bundle
fi

# Make the public stuff public

chmod 644 "${SPIFFE_TRUSTBUNDLE_PATH}"
chmod 755 `dirname "${SPIFFE_TRUSTBUNDLE_PATH}"`
chmod 755 `dirname "${SPIFFE_ENDPOINTSOCKET}"`

# Process agent configuration template

CONF_FILE="/srv/spiffe/agent/agent.conf"

cp -fp /usr/local/etc/spire/agent.conf.tpl "${CONF_FILE}"
sed -i -e "s~SPIFFE_ENDPOINTSOCKET~${SPIFFE_ENDPOINTSOCKET}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_SERVER_SOCKET~${SPIFFE_SERVER_SOCKET}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_TRUSTBUNDLE_PATH~${SPIFFE_TRUSTBUNDLE_PATH}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_TRUSTDOMAIN~${SPIFFE_TRUSTDOMAIN}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_SERVER_HOST~${SPIFFE_SERVER_HOST}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_SERVER_PORT~${SPIFFE_SERVER_PORT}~" "${CONF_FILE}"

# Start the agent

exec spire-agent run -config "${CONF_FILE}"
