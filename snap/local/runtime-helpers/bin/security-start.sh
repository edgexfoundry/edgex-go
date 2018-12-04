#!/bin/bash -e

# the kong wrapper script from $SNAP
export KONG_SNAP="${SNAP}/bin/kong-wrapper.sh"

# general config dirs for security-api-gateway and security-secret-store 
export CONFIG_DIR=${SNAP_DATA}/config
export SEC_SEC_STORE_CONFIG_DIR=${CONFIG_DIR}/security-secret-store
export SEC_API_GATEWAY_CONFIG_DIR=${CONFIG_DIR}/security-api-gateway

# security-secret-store environment variables
export _VAULT_DIR=${SNAP_DATA}/vault
export _PKI_SETUP_VAULT_FILE=${SEC_SEC_STORE_CONFIG_DIR}/pkisetup-vault.json
export _PKI_SETUP_KONG_FILE=${SEC_SEC_STORE_CONFIG_DIR}/pkisetup-kong.json

# kong environment variables
export LOG_DIR=${SNAP_COMMON}/logs
export KONG_PROXY_ACCESS_LOG=${LOG_DIR}/kong-proxy-access.log
export KONG_ADMIN_ACCESS_LOG=${LOG_DIR}/kong-admin-access.log
export KONG_PROXY_ERROR_LOG=${LOG_DIR}/kong-admin-error.log
export KONG_ADMIN_ERROR_LOG=${LOG_DIR}/kong-admin-error.log
export KONG_ADMIN_LISTEN="0.0.0.0:8001, 0.0.0.0:8444 ssl"

# vault environment variables
export VAULT_ADDR=https://localhost:8200
export VAULT_CONFIG_DIR=${_VAULT_DIR}/config
export VAULT_UI=true
export num_tries=0
export MAX_KONG_UP_TRIES=10
export MAX_VAULT_UNSEAL_TRIES=10

# before running anything else go into a snap writable directory
# this prevents issues later on with popd, as we may not have permission to go back to the directory we came from
# on some systems. See issue #509 for more details
cd $SNAP_DATA

# start up cassandra
$SNAP/bin/cassandra-wrapper.sh

# touch all the kong log files to ensure they exist
mkdir -p ${LOG_DIR}
for log in ${KONG_PROXY_ACCESS_LOG} ${KONG_ADMIN_ACCESS_LOG} ${KONG_PROXY_ERROR_LOG} ${KONG_ADMIN_ERROR_LOG}; do
    touch $log
done

# run kong migrations up to bootstrap the cassandra database
# note that sometimes cassandra can in a "starting up" start, etc.
# and in this case we should just loop and keep trying
until $KONG_SNAP migrations up --yes --conf ${SEC_API_GATEWAY_CONFIG_DIR}/kong.conf; do
    sleep 10
    # increment number of tries
    num_tries=$((num_tries+1))
    if (( num_tries > MAX_KONG_UP_TRIES )); then
        echo "max tries attempting to bring up kong"
        exit 1
    fi
done

# now start kong normally
$KONG_SNAP start --conf ${SEC_API_GATEWAY_CONFIG_DIR}/kong.conf

# generate tls keys before running vault
pushd ${_VAULT_DIR} > /dev/null
# check to see if the root certificate generated with pkisetup already exists, if so then don't generate new certs
# note that this assumes::
# * that the pkisetup-vault.json file exists and is valid json
# * that if the root ca file still exists then the other certificates still exists
# * that the root ca file name is located at $working_dir/$pki_setup_dir/$ca_name/$ca_name.pem (this is true for the current release)
CERT_DIR=$(jq -r '.working_dir' "${_PKI_SETUP_VAULT_FILE}")
CERT_SUBDIR=$(jq -r '.pki_setup_dir' "${_PKI_SETUP_VAULT_FILE}")
ROOT_NAME=$(jq -r '.x509_root_ca_parameters | .ca_name' "${_PKI_SETUP_VAULT_FILE}")
if [ ! -f "${CERT_DIR}/${CERT_SUBDIR}/${ROOT_NAME}/${ROOT_NAME}.pem" ]; then
     ${SNAP}/bin/pkisetup --config ${_PKI_SETUP_VAULT_FILE}
     ${SNAP}/bin/pkisetup --config ${_PKI_SETUP_KONG_FILE}
fi
popd > /dev/null

# wait for consul to come up before starting vault
$SNAP/bin/wait-for-consul.sh "security-services"

# execute the vault binary in the background, logging 
# all output to $SNAP_COMMON
$SNAP/bin/vault server \
     --config="${SEC_SEC_STORE_CONFIG_DIR}/vault-config.json" \
     | tee ${LOG_DIR}/vault.log &

# run the vault-worker
pushd ${SEC_SEC_STORE_CONFIG_DIR} > /dev/null
echo "running vault-worker from security-secret-store"
$SNAP/bin/vault-worker --init=true --configfile=${SEC_SEC_STORE_CONFIG_DIR}/res/configuration.toml
# copy the kong access token to the config directory for the security-api-gateway so it has 
# perms to read the certs from vault and upload them into kong
cp ${SEC_SEC_STORE_CONFIG_DIR}/res/kong-token.json ${SEC_API_GATEWAY_CONFIG_DIR}/res/kong-token.json
popd > /dev/null

# now finally start the security proxy
pushd ${SEC_API_GATEWAY_CONFIG_DIR} > /dev/null
echo "running edgexproxy from security-api-gateway"
$SNAP/bin/edgexproxy --configfile=${SEC_API_GATEWAY_CONFIG_DIR}/res/configuration.toml --init=true
popd > /dev/null

# wait for the forked processes to exit
# this is necessary because currently the security-service is implemented as a "notify" daemon which 
# means that systemd will kill the child processes of this process when this process exits
# the ideal thing to do here would be have this be a "forking" daemon, but that has problems
# timing out waiting for the fork to happen, as we need to wait for cassandra to finish coming up
# before the default systemd service start timeout of 30 seconds
wait

