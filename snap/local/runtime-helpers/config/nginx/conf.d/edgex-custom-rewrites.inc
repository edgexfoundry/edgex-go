# Add custom location directives to this file, for example:

location /device-virtual {
  rewrite            /device-virtual/(.*) /$1 break;
  proxy_pass         http://127.0.0.1:59900;
  proxy_redirect     off;
  proxy_set_header   Host $host;
  auth_request       /auth;
  auth_request_set   $auth_status $upstream_status;
}
