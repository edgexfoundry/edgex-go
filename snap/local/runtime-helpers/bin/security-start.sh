#!/bin/bash -e

export KONG_SNAP="${SNAP}/bin/kong-wrapper.sh"

export LOG_DIR=${SNAP_COMMON}/logs
export CONFIG_DIR=${SNAP_DATA}/config
export SEC_SEC_STORE_CONFIG_DIR=${CONFIG_DIR}/security-secret-store
export SEC_API_GATEWAY_CONFIG_DIR=${CONFIG_DIR}/security-api-gateway

# security-secret-store environment variables
export _VAULT_SCRIPT_DIR=${SNAP}/bin
export _VAULT_DIR=${SNAP_DATA}/vault
export _VAULT_SVC=localhost
export _KONG_SVC=localhost
export _PKI_SETUP_VAULT_ENV=${SEC_SEC_STORE_CONFIG_DIR}/pki-setup-config-vault.env
export _PKI_SETUP_KONG_ENV=${SEC_SEC_STORE_CONFIG_DIR}/pki-setup-config-kong.env
export WATCHDOG_DELAY=3m

# security-api-gateway environment variables
export KONG_PROXY_ACCESS_LOG=${LOG_DIR}/kong-proxy-access.log
export KONG_ADMIN_ACCESS_LOG=${LOG_DIR}/kong-admin-access.log
export KONG_PROXY_ERROR_LOG=${LOG_DIR}/kong-admin-error.log
export KONG_ADMIN_ERROR_LOG=${LOG_DIR}/kong-admin-error.log
export KONG_ADMIN_LISTEN="0.0.0.0:8001, 0.0.0.0:8444 ssl"
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

# setup key generation for vault before starting vault up
# note that the vault setup scripts will put the generated keys inside
# ./pki/... so we go into the $SNAP_DATA/vault which is where all the other 
# scripts go. Ideally we would override some environment variables and not have
# to change directories for this
pushd ${_VAULT_DIR}
$SNAP/bin/vault-setup.sh
popd

# wait for consul to come up before starting vault
$SNAP/bin/wait-for-consul.sh "security-services"

# execute the vault binary in the background, logging 
# all output to $SNAP_COMMONG
$SNAP/bin/vault server \
     --config="${SEC_SEC_STORE_CONFIG_DIR}/vault-config.json" \
     | tee ${LOG_DIR}/vault.log &

# now that vault is up and running we need to initialize it
# same situation as for vault when setting up - we need to be in the vault folder
# for this to work
pushd ${_VAULT_DIR}
while true; do
    # Init/Unseal processes
    set +e
    ${_VAULT_SCRIPT_DIR}/vault-init-unseal.sh
    init_res=$?
    set -e
    
    # If Vault init/unseal was OK break out of the loop
    # to start the vault-worker running in the background indefinitely
    if [ "$init_res" -eq 0 ];  then
        break
    fi

    # increment number of tries
    num_tries=$((num_tries+1))
    if (( num_tries > MAX_VAULT_UNSEAL_TRIES )); then
        echo "max tries attempting to unseal vault"
        exit 1
    fi

    # we sleep for 1 second each time in this first loop to check that it quickly comes up
    # the background loop sleeps for 3 minutes each iteration
    sleep 1
done

# now that we are done unsealing the vault, we can start the 
# vault-worker in the background
# NOTE: for Delhi we want to change this to not just always run and 
# have a better restarting mechanism in case vault goes down
# see discussion on https://github.com/edgexfoundry/security-secret-store/pull/15#issuecomment-412043938
# also note that the process vault-init-unseal will probably immediately get run again here, but that's
# fine, as there's no harm in unsealing vault when it's already unsealed
$SNAP/bin/vault-worker.sh &
popd

# go into the snap folder
cd ${_VAULT_DIR}/file

# copy the resp-init.json file into the res folder
mkdir -p res
cp resp-init.json res/

# now finally start the security proxy
$SNAP/bin/edgexproxy --configfile=${SEC_API_GATEWAY_CONFIG_DIR}/res/configuration.toml init=true
