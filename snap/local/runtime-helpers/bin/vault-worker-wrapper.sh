#!/bin/bash -e

export CONFIG_DIR=${SNAP_DATA}/config
export SEC_SEC_STORE_CONFIG_DIR=${CONFIG_DIR}/security-secret-store
export SEC_API_GATEWAY_CONFIG_DIR=${CONFIG_DIR}/security-api-gateway

# run the vault-worker
cd ${SEC_SEC_STORE_CONFIG_DIR}

$SNAP/bin/vault-worker --init=true --configfile=${SEC_SEC_STORE_CONFIG_DIR}/res/configuration.toml
# copy the kong access token to the config directory for the security-api-gateway so it has 
# perms to read the certs from vault and upload them into kong
cp ${SEC_SEC_STORE_CONFIG_DIR}/res/kong-token.json ${SEC_API_GATEWAY_CONFIG_DIR}/res/kong-token.json
