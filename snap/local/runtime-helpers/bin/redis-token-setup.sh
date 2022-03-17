#!/bin/bash

logger "edgexfoundry.kuiper:redis-token-setup: started"

VAULT_TOKEN_FILE=$SNAP_DATA/secrets/edgex-ekuiper/secrets-token.json
SOURCE_FILE=$SNAP_DATA/kuiper/etc/sources/edgex.yaml 
CONNECTIONS_FILE=$SNAP_DATA/kuiper/etc/connections/connection.yaml

handle_error()
{
    local EXIT_CODE=$1
    local ITEM=$2
    local RESPONSE=$3
    if [ $EXIT_CODE -ne 0 ] ; then
        logger --stderr "edgexfoundry.kuiper:redis-token-setup: $ITEM exited with code $EXIT_CODE: $RESPONSE"
        exit 1
    fi
}

# use Vault token query Redis token, access edgexfoundry secure Message Bus
if [ -f "$VAULT_TOKEN_FILE" ] ; then
    # get Vault token and generate Redis credentials
    logger "edgexfoundry.kuiper:redis-token-setup: using Vault token to query Redis token"
    TOKEN=$(yq "$VAULT_TOKEN_FILE" | yq ' .auth.client_token')
    handle_error $? "yq" $TOKEN

    # check CURL's exit code
    CURL_RES=$(curl --silent --write-out "%{http_code}" \
    --header "X-Vault-Token: $TOKEN" \
    --request GET http://localhost:8200/v1/secret/edgex/edgex-ekuiper/redisdb)
    handle_error $? "curl" $CURL_RES

    # check response http code
    HTTP_CODE="${CURL_RES:${#CURL_RES}-3}"
    if [ $HTTP_CODE -ne 200 ] ; then
        logger --stderr "edgexfoundry.kuiper:redis-token-setup: http error $HTTP_CODE, with response: $CURL_RES"
        exit 1
    fi

    # get CURL's reponse
    if [ ${#CURL_RES} -eq 3 ]; then
        logger --stderr "edgexfoundry.kuiper:redis-token-setup: unexpected http response with empty body"
        exit 1
    else
        BODY="${CURL_RES:0:${#CURL_RES}-3}"
    fi

    # process the reponse and check if yq works
    REDIS_USER=$(echo $BODY| yq '.data.username')
    handle_error $? "yq" $REDIS_USER
    REDIS_PASS=$(echo $BODY| yq '.data.password')
    handle_error $? "yq" $REDIS_PASS

    # pass generated Redis credentials to configuration files
    logger "edgexfoundry.kuiper:redis-token-setup: adding Redis credentials to $SOURCE_FILE"
    YQ_RES=$(yq -i '.default += {"optional":{"Username":"'$REDIS_USER'"}+{"Password":"'$REDIS_PASS'"}}' "$SOURCE_FILE")
    handle_error $? "yq" $YQ_RES
    
    logger "edgexfoundry.kuiper:redis-token-setup: adding Redis credentials to $CONNECTIONS_FILE"
    YQ_RES=$(yq -i '.edgex.redisMsgBus += {"optional":{"Username":"'$REDIS_USER'"}+{"Password":"'$REDIS_PASS'"}}' "$CONNECTIONS_FILE")
    handle_error $? "yq" $YQ_RES

    logger "edgexfoundry.kuiper:redis-token-setup: configured Kuiper to authenticate with Redis, using credentials fetched from Vault"
else
    logger --stderr "edgexfoundry.kuiper:redis-token-setup: unable to configure Kuiper to authenticate with Redis: unable to query Redis token from Vault: Vault token not available"
fi

exec "$@"