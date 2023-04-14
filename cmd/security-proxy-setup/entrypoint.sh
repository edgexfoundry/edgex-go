#!/usr/bin/dumb-init /bin/sh
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

set -e

keyfile=nginx.key
certfile=nginx.crt

# This script should run as root as it contains chown commands.
# SKIP_CHOWN is provided in the event that it can be arranged
# that /etc/ssl/nginx is writable by uid/gid 101
# and the container is also started as uid/gid 101.

# Check for default TLS certificate for reverse proxy, create if missing
# Normally we would run the below command in the nginx container itself,
# but nginx:alpine-slim does not include openssl, thus run it here instead.
if test -d /etc/ssl/nginx ; then
    cd /etc/ssl/nginx
    if test ! -f "${keyfile}" ; then
        # (NGINX will restart in a failure loop until a TLS key exists)
        # Create default TLS certificate with 1 day expiry -- user must replace in production (do this as nginx user)
        openssl req -x509 -nodes -days 1 -newkey ec -pkeyopt ec_paramgen_curve:secp384r1 -subj '/CN=localhost/O=EdgeX Foundry' -keyout "${keyfile}" -out "${certfile}" -addext "keyUsage = digitalSignature, keyCertSign" -addext "extendedKeyUsage = serverAuth"
        if [ -z "$SKIP_CHOWN" ]; then
            # nginx process user is 101:101
            chown 101:101 "${keyfile}" "${certfile}"
        fi
        echo "Default TLS certificate created.  Recommend replace with your own."
    fi
fi

#
# Generate custom forwarders based on EDGEX_ADD_PROXY_ROUTE
#

# Truncate the template file before we start appending
: >/etc/nginx/templates/generated-routes.inc.template

IFS=', '
for service in ${EDGEX_ADD_PROXY_ROUTE}; do
	prefix=$(echo -n "${service}" | sed -n -e 's/\([-0-9a-zA-Z]*\)\..*/\1/p')
	host=$(echo -n "${service}" | sed -n -e 's/.*\/\/\([-0-9a-zA-Z]*\):.*/\1/p')
	port=$(echo -n "${service}" | sed -n -e 's/.*:\(\d*\)/\1/p')
	varname=$(echo -n "${prefix}" | tr '-' '_')
	echo $service $prefix $host $port
  cat >> /etc/nginx/templates/generated-routes.inc.template <<EOH

set \$upstream_$varname $host;
location /$prefix {
  rewrite            /$prefix/(.*) /\$1 break;
  resolver           127.0.0.11 valid=30s;
  proxy_pass         http://\$upstream_$varname:$port;
  proxy_redirect     off;
  proxy_set_header   Host \$host;
  auth_request       /auth;
  auth_request_set   \$auth_status \$upstream_status;
}
EOH

done
unset IFS



# Hang the container now that initialization is done.
cd /
exec su nobody -s /bin/sh -c "exec tail -f /dev/null"
