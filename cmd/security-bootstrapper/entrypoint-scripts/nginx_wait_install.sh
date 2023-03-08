#!/bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2023 Intel Corporation
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

# env settings are populated from env files of docker-compose

echo "Awaiting ReadyToRun signal prior to starting NGINX"

# gating on the ready-to-run port
echo "$(date) Executing waitFor with waiting on tcp://${STAGEGATE_BOOTSTRAPPER_HOST}:${STAGEGATE_READY_TORUNPORT}"
/edgex-init/security-bootstrapper --configDir=/edgex-init/res waitFor \
  -uri tcp://"${STAGEGATE_BOOTSTRAPPER_HOST}":"${STAGEGATE_READY_TORUNPORT}" \
  -timeout "${STAGEGATE_WAITFOR_TIMEOUT}"


echo "$(date) Generating default config ..."

# Ensure this file exists since reference below; proxy-setup will regenerate it
touch /etc/nginx/templates/generated-routes.inc.template

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

cat <<'EOH' > /etc/nginx/templates/edgex-default.conf.template
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

    set $upstream_proxyauth edgex-proxy-auth;
    location /auth {
      internal;
      resolver                127.0.0.11 valid=30s;
      proxy_pass              http://$upstream_proxyauth:59842;
      proxy_redirect          off;
      proxy_set_header        Host $host;
      proxy_set_header        Content-Length "";
      proxy_set_header        X-Forwarded-URI $request_uri;
      proxy_pass_request_body off;
    }

    # Rewriting rules (variable usage required to avoid nginx crash if host not resolveable at time of boot)
    # resolver required to enable name resolution at runtime, points at docker DNS resolver

    set $upstream_core_data edgex-core-data;
    location /core-data {
      rewrite            /core-data/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_core_data:59880;
      proxy_redirect     off;
      proxy_set_header   Host $host;
      auth_request       /auth;
      auth_request_set   $auth_status $upstream_status;
    }


    set $upstream_core_metadata edgex-core-metadata;
    location /core-metadata {
      rewrite            /core-metadata/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_core_metadata:59881;
      proxy_redirect     off;
      proxy_set_header   Host $host;
      auth_request       /auth;
      auth_request_set   $auth_status $upstream_status;
    }


    set $upstream_core_command edgex-core-command;
    location /core-command {
      rewrite            /core-command/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_core_command:59882;
      proxy_redirect     off;
      proxy_set_header   Host $host;
      auth_request       /auth;
      auth_request_set   $auth_status $upstream_status;
    }


    set $upstream_support_notifications edgex-support-notifications;
    location /support-notifications {
      rewrite            /support-notifications/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_support_notifications:59860;
      proxy_redirect     off;
      proxy_set_header   Host $host;
      auth_request       /auth;
      auth_request_set   $auth_status $upstream_status;
    }


    set $upstream_support_scheduler edgex-support-scheduler;
    location /support-scheduler {
      rewrite            /support-scheduler/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_support_scheduler:59861;
      proxy_redirect     off;
      proxy_set_header   Host $host;
      auth_request       /auth;
      auth_request_set   $auth_status $upstream_status;
    }

    set $upstream_app_rules_engine edgex-app-rules-engine;
    location /app-rules-engine {
      rewrite            /app-rules-engine/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_app_rules_engine:59701;
      proxy_redirect     off;
      proxy_set_header   Host $host;
      auth_request       /auth;
      auth_request_set   $auth_status $upstream_status;
    }

    set $upstream_kuiper edgex-kuiper;
    location /rules-engine {
      rewrite            /rules-engine/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_kuiper:59720;
      proxy_redirect     off;
      proxy_set_header   Host $host;
      auth_request       /auth;
      auth_request_set   $auth_status $upstream_status;
    }

    set $upstream_device_virtual edgex-device-virtual;
    location /device-virtual {
      rewrite            /device-virtual/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_device_virtual:59900;
      proxy_redirect     off;
      proxy_set_header   Host $host;
      auth_request       /auth;
      auth_request_set   $auth_status $upstream_status;
    }

    # Note: Consul implements its own authentication mechanism (only allow API, /v1, through)
    set $upstream_core_consul edgex-core-consul;
    location /consul/v1 {
      rewrite            /consul/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_core_consul:8500;
      proxy_redirect     off;
      proxy_set_header   Host $host;
    }

    # Note: Vault login API does not require authentication at the gateway for obvious reasons
    set $upstream_vault edgex-vault;
    location /vault/v1/auth/userpass/login {
      rewrite            /vault/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_vault:8200;
      proxy_redirect     off;
      proxy_set_header   Host $host;
    }
    location /vault/v1/identity/oidc/token {
      rewrite            /vault/(.*) /$1 break;
      resolver           127.0.0.11 valid=30s;
      proxy_pass         http://$upstream_vault:8200;
      proxy_redirect     off;
      proxy_set_header   Host $host;
    }

    include /etc/nginx/conf.d/generated-routes.inc;
    include /etc/nginx/conf.d/edgex-custom-rewrites.inc;

}

# Don't output NGINX version in Server: header
server_tokens off;

EOH

echo "$(date) Starting NGINX ..."
exec /docker-entrypoint.sh "$@"
