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

for dir in `dirname "${SPIFFE_SERVER_SOCKET}"` \
           `dirname "${SPIFFE_ENDPOINTSOCKET}"` \
           /srv/spiffe/ca/public \
           /srv/spiffe/ca/private ; do
    test -d "$dir" || mkdir -p "$dir"
done

# CA SPIFFE identifiers

if test ! -f "/srv/spiffe/ca/public/ca.crt"; then
    openssl ecparam -genkey -name secp521r1 -noout -out "/srv/spiffe/ca/private/ca.key"
    SAN="" openssl req -subj "/CN=SPIFFE Root CA" -config "/usr/local/etc/openssl.conf" -key "/srv/spiffe/ca/private/ca.key" -sha512 -new -out "/run/ca.req.$$"
    SAN="" openssl x509 -sha512 -signkey "/srv/spiffe/ca/private/ca.key" -clrext -extfile /usr/local/etc/openssl.conf -extensions ca_ext -CAkey "/srv/spiffe/ca/private/ca.key" -CAcreateserial -req -in "/run/ca.req.$$" -days 3650 -out "/srv/spiffe/ca/public/ca.crt"
    rm -f "/run/ca.req.$$"
fi

# CA for node (agent) attestation

if test ! -f "/srv/spiffe/ca/public/agent-ca.crt"; then
    openssl ecparam -genkey -name secp521r1 -noout -out "/srv/spiffe/ca/private/agent-ca.key"
    SAN="" openssl req -subj "/CN=SPIFFE Agent CA" -config "/usr/local/etc/openssl.conf" -key "/srv/spiffe/ca/private/agent-ca.key" -sha512 -new -out "/run/ca.req.$$"
    SAN="" openssl x509 -sha512 -signkey "/srv/spiffe/ca/private/agent-ca.key" -clrext -extfile /usr/local/etc/openssl.conf -extensions ca_ext -CAkey "/srv/spiffe/ca/private/agent-ca.key" -CAcreateserial -req -in "/run/ca.req.$$" -days 3650 -out "/srv/spiffe/ca/public/agent-ca.crt"
    rm -f "/run/ca.req.$$"
fi

# Process server configuration template

CONF_FILE="/srv/spiffe/server/server.conf"
cp -fp /usr/local/etc/spire/server.conf.tpl "${CONF_FILE}"

sed -i -e "s~SPIFFE_ENDPOINTSOCKET~${SPIFFE_ENDPOINTSOCKET}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_SERVER_SOCKET~${SPIFFE_SERVER_SOCKET}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_TRUSTBUNDLE_PATH~${SPIFFE_TRUSTBUNDLE_PATH}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_TRUSTDOMAIN~${SPIFFE_TRUSTDOMAIN}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_SERVER_HOST~${SPIFFE_SERVER_HOST}~" "${CONF_FILE}"
sed -i -e "s~SPIFFE_SERVER_PORT~${SPIFFE_SERVER_PORT}~" "${CONF_FILE}"

exec spire-server run -config "${CONF_FILE}"
