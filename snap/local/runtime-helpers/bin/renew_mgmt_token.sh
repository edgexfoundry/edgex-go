#!/bin/bash -e
#
# This is a service using a shell script for renewing the Vault's management token that is used to generate the Consul tokens
# for EdgeX services.  The life cycle of those Consul tokens are governed by Vault and hence if the Vault token
# is expired, then all these Consul tokens will be revoked from Consul as well. In order for EdgeX services keeping
# using the valid Consul tokens, we need to renew the management token periodically while it is running.
#

# function to self-renew Consul secrets engine's Vault management token
# the self-renew occurs after its TTL < 10% of its period
renewToken()
{
    mgmt_token="$1"

    lookup_url="http://${SECRETSTORE_HOST}:${SECRETSTORE_PORT}/v1/auth/token/lookup-self"
    renew_url="http://${SECRETSTORE_HOST}:${SECRETSTORE_PORT}/v1/auth/token/renew-self"

    # loop forever for token-renewal process
    while true; do
        echo "checking Vault management token's TTL with ${lookup_url} ..."
        # self lookup token's period and remaining life (ttl)
        response=$(curl -s -w "__STATUS_CODE__=%{http_code}" -H "Cache-Control: no-cache" \
            -H "X-Vault-Token:${mgmt_token}" "${lookup_url}")
        status_code="${response#*__STATUS_CODE__=}"
        json_data="${response%__STATUS_CODE__=*}"
        if [ "${status_code}" -eq 200 ]; then
            # caluclate the amount of time to sleep before renew the token
            # the self-renew will occur when its TTL < 10% of its period
            wait_time=$(echo "${json_data}" | jq -r '.data.ttl-.data.period/10')
            if [ "${wait_time}" -gt 0 ]; then
                echo "management token will be renewed after ${wait_time} seconds"
                sleep "${wait_time}"
            else
                # alredy below the threshold: no wait, just renew it if hasn't expired yet
                echo "management token's TTL is lower than 10%"
            fi
            echo "try to renew management token with ${renew_url}:"
            renew_status_code=$(curl -X POST -s -o /dev/null -w "%{http_code}" -H "X-Vault-Token:${mgmt_token}" "${renew_url}")
            if [ "${renew_status_code}" -eq 200 ]; then
              echo "  successfully renewed management token"
            else
              echo "  ERROR: unable to renew management token, status code = ${renew_status_code}"
            fi
        else
            echo "ERROR: failed to lookup-self for management token due to token already expired or insufficient permission, cannot renew it. Please re-generate a new management token."
            return 1
        fi
    done

    echo "renewToken process done"
}

echo "in renew_mgmt_token.sh: ENABLE_REGISTRY_ACL = ${ENABLE_REGISTRY_ACL}"

if [ "${ENABLE_REGISTRY_ACL}" == "true" ]; then
    # kick off the token renewal process
    echo "launching vault management token renewal process:"
    if [ -f "${STAGEGATE_REGISTRY_ACL_SECRETSADMINTOKENPATH}" ]; then
        mgmt_token=$("${SNAP}"/usr/bin/jq -r '.auth.client_token' "${STAGEGATE_REGISTRY_ACL_SECRETSADMINTOKENPATH}")
        parsing_token_ret_code=$?
        if [ "${parsing_token_ret_code}" -eq 0 ]; then
            renewToken "${mgmt_token}"
        else
            "ERROR: cannot parse token from file ${STAGEGATE_REGISTRY_ACL_SECRETSADMINTOKENPATH}, unexpected JSON schema"
        fi
    else
        echo "ERROR: token file ${STAGEGATE_REGISTRY_ACL_SECRETSADMINTOKENPATH} does not exist, cannot renew token"
    fi
else
    echo "Consul ACL not enabled, skip renewing management token"
fi
