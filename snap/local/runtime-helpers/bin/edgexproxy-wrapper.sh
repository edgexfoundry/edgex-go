#!/bin/bash -e

export SEC_API_GATEWAY_CONFIG_DIR=${SNAP_DATA}/config/security-api-gateway

cd ${SEC_API_GATEWAY_CONFIG_DIR}
$SNAP/bin/edgexproxy --configfile=${SEC_API_GATEWAY_CONFIG_DIR}/res/configuration.toml --init=true
