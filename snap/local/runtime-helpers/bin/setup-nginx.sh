#!/bin/bash

keyfile=nginx.key
certfile=nginx.crt

# Check for default TLS certificate for reverse proxy, create if missing
# Normally we would run the below command in the nginx container itself,
# but nginx:alpine-slim does not container openssl, thus run it here instead.
mkdir -p "${SNAP_DATA}/nginx"
cd "${SNAP_DATA}/nginx"
if test ! -f "${keyfile}" ; then
    # (NGINX will restart in a failure loop until a TLS key exists)
    # Create default TLS certificate with 1 day expiry -- user must replace in production (do this as nginx user)
    openssl req -x509 -nodes -days 1 -newkey ec -pkeyopt ec_paramgen_curve:secp384r1 -subj '/CN=localhost/O=EdgeX Foundry' -keyout "${keyfile}" -out "${certfile}" -addext "keyUsage = digitalSignature, keyCertSign" -addext "extendedKeyUsage = serverAuth"
    echo "Default TLS certificate created.  Recommend replace with your own."
fi
