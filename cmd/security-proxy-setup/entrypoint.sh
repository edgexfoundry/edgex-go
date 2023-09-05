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
# Import CORS configuration from common config
#

: ${EDGEX_SERVICE_CORSCONFIGURATION_ENABLECORS:=`yq -r .all-services.Service.CORSConfiguration.EnableCORS /edgex/res/common_configuration.yaml`}
: ${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWCREDENTIALS:=`yq -r .all-services.Service.CORSConfiguration.CORSAllowCredentials /edgex/res/common_configuration.yaml`}
: ${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWEDORIGIN:=`yq -r .all-services.Service.CORSConfiguration.CORSAllowedOrigin /edgex/res/common_configuration.yaml`}
: ${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWEDMETHODS:=`yq -r .all-services.Service.CORSConfiguration.CORSAllowedMethods /edgex/res/common_configuration.yaml`}
: ${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWEDHEADERS:=`yq -r .all-services.Service.CORSConfiguration.CORSAllowedHeaders /edgex/res/common_configuration.yaml`}
: ${EDGEX_SERVICE_CORSCONFIGURATION_CORSEXPOSEHEADERS:=`yq -r .all-services.Service.CORSConfiguration.CORSExposeHeaders /edgex/res/common_configuration.yaml`}
: ${EDGEX_SERVICE_CORSCONFIGURATION_CORSMAXAGE:=`yq -r .all-services.Service.CORSConfiguration.CORSMaxAge /edgex/res/common_configuration.yaml`}

echo "$(date) CORS settings dump ..."
( set | grep EDGEX_SERVICE_CORSCONFIGURATION ) || true

# See https://github.com/edgexfoundry/edgex-go/issues/4648 as to why CORS is implemented this way.
# Warning: no not simplify add_header redundancy. See https://www.peterbe.com/plog/be-very-careful-with-your-add_header-in-nginx
corssnippet=/etc/nginx/templates/cors.block.$$
touch "${corssnippet}"
if test "${EDGEX_SERVICE_CORSCONFIGURATION_ENABLECORS}" = "true"; then
  echo "      if (\$request_method = 'OPTIONS') {" >> "${corssnippet}"
  echo "        add_header 'Access-Control-Allow-Origin' '${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWEDORIGIN}';" >> "${corssnippet}"
  echo "        add_header 'Access-Control-Allow-Methods' '${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWEDMETHODS}';" >> "${corssnippet}"
  echo "        add_header 'Access-Control-Allow-Headers' '${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWEDHEADERS}';" >> "${corssnippet}"
  if test "${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWCREDENTIALS}" = "true"; then
    # CORS specificaiton says that if not true, omit the header entirely
    echo "        add_header 'Access-Control-Allow-Credentials' '${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWCREDENTIALS}';" >> "${corssnippet}"
  fi
  echo "        add_header 'Access-Control-Max-Age' ${EDGEX_SERVICE_CORSCONFIGURATION_CORSMAXAGE};" >> "${corssnippet}"
  echo "        add_header 'Vary' 'origin';" >> "${corssnippet}"
  echo "        add_header 'Content-Type' 'text/plain; charset=utf-8';" >> "${corssnippet}"
  echo "        add_header 'Content-Length' 0;" >> "${corssnippet}"
  echo "        return 204;" >> "${corssnippet}"
  echo "      }" >> "${corssnippet}"
  echo "      if (\$request_method != 'OPTIONS') {" >> "${corssnippet}"
  # Always add headers regardless of response code.  Omit preflight-related headers (allow-methods, allow-headers, allow-credentials, max-age)
  echo "        add_header 'Access-Control-Allow-Origin' '${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWEDORIGIN}' always;" >> "${corssnippet}"
  echo "        add_header 'Access-Control-Expose-Headers' '${EDGEX_SERVICE_CORSCONFIGURATION_CORSEXPOSEHEADERS}' always;" >> "${corssnippet}"
  if test "${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWCREDENTIALS}" = "true"; then
    # CORS specificaiton says that if not true, omit the header entirely
    echo "        add_header 'Access-Control-Allow-Credentials' '${EDGEX_SERVICE_CORSCONFIGURATION_CORSALLOWCREDENTIALS}';" >> "${corssnippet}"
  fi
  echo "        add_header 'Vary' 'origin' always;" >> "${corssnippet}"
  echo "      }" >> "${corssnippet}"
  echo "" >> "${corssnippet}"
fi

#
# Generate NGINX configuration based on EDGEX_ADD_PROXY_ROUTE and standard settings
#

echo "$(date) Generating default NGINX config ..."

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
`cat "${corssnippet}"`
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


# This file can be modified by the user; deleted when docker volumes are pruned;
# but preserved across start/up and stop/down actions
if test -f /etc/nginx/templates/edgex-custom-rewrites.inc.template; then
  echo "Using existing custom-rewrites."
else
  cat <<'EOH' > /etc/nginx/templates/edgex-custom-rewrites.inc.template
# Add custom location directives to this file, for example:

# set $upstream_device_virtual edgex-device-virtual;
# location /device-virtual {
#   rewrite            /device-virtual/(.*) /$1 break;
#   resolver           127.0.0.11 valid=30s;
#   proxy_pass         http://$upstream_device_virtual:59900;
#   proxy_redirect     off;
#   proxy_set_header   Host $host;
#   auth_request       /auth;
#   auth_request_set   $auth_status $upstream_status;
# }
EOH
fi

cat <<EOH > /etc/nginx/templates/edgex-default.conf.template
#
# Copyright (C) Intel Corporation 2023
# SPDX-License-Identifier: Apache-2.0
#

# generated 2023-01-19, Mozilla Guideline v5.6, nginx 1.17.7, OpenSSL 1.1.1k, modern configuration, no HSTS, no OCSP
# https://ssl-config.mozilla.org/#server=nginx&version=1.17.7&config=modern&openssl=1.1.1k&hsts=false&ocsp=false&guideline=5.6
server {
    listen 8000;  # Docker deployments only
    listen 8443 ssl;

    ssl_certificate "/etc/ssl/nginx/nginx.crt";
    ssl_certificate_key "/etc/ssl/nginx/nginx.key";
    ssl_session_tickets off;
    # modern configuration
    ssl_protocols TLSv1.3;
    ssl_prefer_server_ciphers off;


    # Subrequest authentication

    set \$upstream_proxyauth edgex-proxy-auth;
    location /auth {
      internal;
      resolver                127.0.0.11 valid=30s;
      proxy_pass              http://\$upstream_proxyauth:59842;
      proxy_redirect          off;
      proxy_set_header        Host \$host;
      proxy_set_header        Content-Length "";
      proxy_set_header        X-Forwarded-URI \$request_uri;
      proxy_pass_request_body off;
    }

    # Rewriting rules (variable usage required to avoid nginx crash if host not resolveable at time of boot)
    # resolver required to enable name resolution at runtime, points at docker DNS resolver

    set \$upstream_core_data edgex-core-data;
    location /core-data {
`cat "${corssnippet}"`
      rewrite            /core-data/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_core_data:59880;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
      auth_request       /auth;
      auth_request_set   \$auth_status \$upstream_status;
    }


    set \$upstream_core_metadata edgex-core-metadata;
    location /core-metadata {
`cat "${corssnippet}"`
      rewrite            /core-metadata/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_core_metadata:59881;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
      auth_request       /auth;
      auth_request_set   \$auth_status \$upstream_status;
    }


    set \$upstream_core_command edgex-core-command;
    location /core-command {
`cat "${corssnippet}"`
      rewrite            /core-command/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_core_command:59882;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
      auth_request       /auth;
      auth_request_set   \$auth_status \$upstream_status;
    }


    set \$upstream_support_notifications edgex-support-notifications;
    location /support-notifications {
`cat "${corssnippet}"`
      rewrite            /support-notifications/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_support_notifications:59860;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
      auth_request       /auth;
      auth_request_set   \$auth_status \$upstream_status;
    }


    set \$upstream_support_scheduler edgex-support-scheduler;
    location /support-scheduler {
`cat "${corssnippet}"`
      rewrite            /support-scheduler/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_support_scheduler:59861;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
      auth_request       /auth;
      auth_request_set   \$auth_status \$upstream_status;
    }

    set \$upstream_app_rules_engine edgex-app-rules-engine;
    location /app-rules-engine {
`cat "${corssnippet}"`
      rewrite            /app-rules-engine/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_app_rules_engine:59701;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
      auth_request       /auth;
      auth_request_set   \$auth_status \$upstream_status;
    }

    set \$upstream_kuiper edgex-kuiper;
    location /rules-engine {
`cat "${corssnippet}"`
      rewrite            /rules-engine/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_kuiper:59720;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
      auth_request       /auth;
      auth_request_set   \$auth_status \$upstream_status;
    }

    set \$upstream_device_virtual edgex-device-virtual;
    location /device-virtual {
`cat "${corssnippet}"`
      rewrite            /device-virtual/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_device_virtual:59900;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
      auth_request       /auth;
      auth_request_set   \$auth_status \$upstream_status;
    }

    # Note: Consul implements its own authentication mechanism (only allow API, /v1, through)
    set \$upstream_core_consul edgex-core-consul;
    location /consul/v1 {
`cat "${corssnippet}"`
      rewrite            /consul/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_core_consul:8500;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
    }

    # Note: Vault login API does not require authentication at the gateway for obvious reasons
    set \$upstream_vault edgex-vault;
    location /vault/v1/auth/userpass/login {
`cat "${corssnippet}"`
      rewrite            /vault/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_vault:8200;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
    }
    location /vault/v1/identity/oidc/token {
`cat "${corssnippet}"`
      rewrite            /vault/(.*) /\$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://\$upstream_vault:8200;
      proxy_redirect     off;
      proxy_set_header   Host \$host;
    }

    include /etc/nginx/conf.d/generated-routes.inc;
    include /etc/nginx/conf.d/edgex-custom-rewrites.inc;

}

# Don't output NGINX version in Server: header
server_tokens off;

EOH

rm -f "${corssnippet}"

# Secure entrypoint script will block on opening a TCP listener after this script exits
