#!/bin/bash -e

export KONG_SNAP="${SNAP}/bin/kong-wrapper.sh"

export LOG_DIR=${SNAP_COMMON}/logs
export CONFIG_DIR=${SNAP_DATA}/config
export SEC_SEC_STORE_CONFIG_DIR=${CONFIG_DIR}/security-secret-store
export SEC_GATEWAY_API_CONFIG_DIR=${CONFIG_DIR}/security-gateway-api

# security-secret-store environment variables
export _VAULT_SCRIPT_DIR=${SNAP}/bin
export _VAULT_DIR=${SNAP_DATA}/vault
export _VAULT_SVC=localhost
export _KONG_SVC=localhost
export _PKI_SETUP_VAULT_ENV=${SEC_SEC_STORE_CONFIG_DIR}/pki-setup-config-vault.env
export _PKI_SETUP_KONG_ENV=${SEC_SEC_STORE_CONFIG_DIR}/pki-setup-config-kong.env
export WATCHDOG_DELAY=10s

# security-gateway-api environment variables
export KONG_PROXY_ACCESS_LOG=${LOG_DIR}/kong-proxy-access.log
export KONG_ADMIN_ACCESS_LOG=${LOG_DIR}/kong-admin-access.log
export KONG_PROXY_ERROR_LOG=${LOG_DIR}/kong-admin-error.log
export KONG_ADMIN_ERROR_LOG=${LOG_DIR}/kong-admin-error.log
export KONG_ADMIN_LISTEN="0.0.0.0:8001, 0.0.0.0:8444 ssl"
export VAULT_ADDR=https://localhost:8200
export VAULT_CONFIG_DIR=${_VAULT_DIR}/config
export VAULT_UI=true

# touch all the kong log files to ensure they exist
mkdir -p ${LOG_DIR}
for log in ${KONG_PROXY_ACCESS_LOG} ${KONG_ADMIN_ACCESS_LOG} ${KONG_PROXY_ERROR_LOG} ${KONG_ADMIN_ERROR_LOG}; do
    touch $log
done

# run kong migrations up to bootstrap the cassandra database
$KONG_SNAP migrations up --yes --conf ${SEC_GATEWAY_API_CONFIG_DIR}/kong.conf

# now start kong normally
$KONG_SNAP start --conf ${SEC_GATEWAY_API_CONFIG_DIR}/kong.conf

# setup key generation for vault before starting vault up
# note that the vault setup scripts will put the generated keys inside
# ./pki/... so we go into the $SNAP_DATA/vault which is where all the other 
# scripts go. Ideally we would override some environment variables and not have
# to change directories for this
pushd ${_VAULT_DIR}
$SNAP/bin/vault-setup.sh
popd

# execute the vault binary in the background, logging 
# all output to $SNAP_COMMONG
$SNAP/bin/vault server \
     --config="${SEC_SEC_STORE_CONFIG_DIR}/vault-config.json" \
     | tee ${LOG_DIR}/vault.log &

# now that vault is up and running we need to initialize it
# same situation as for vault when setting up - we need to be in the vault folder
# for this to work
pushd ${_VAULT_DIR}
$SNAP/bin/vault-worker.sh
popd

# go into the snap folder
cd ${_VAULT_DIR}/file

# copy the resp-init.json file into the res folder
mkdir -p res
cp resp-init.json res/

# now finally start the security proxy
$SNAP/bin/edgexproxy --configfile=${SEC_GATEWAY_API_CONFIG_DIR}/res/configuration.toml init=true
