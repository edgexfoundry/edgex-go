#!/usr/bin/dumb-init /bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2019 Intel Corporation
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
#  SPDX-License-Identifier: Apache-2.0'
#  ----------------------------------------------------------------------------------

set -e

# Use dumb-init as PID 1 in order to reap zombie processes and forward system signals to 
# all processes in its session. This can alleviate the chance of leaking zombies, 
# thus more graceful termination of all sub-processes if any.

# runtime directory is set per user:
export XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR:-/run/user/$(echo $(id -u))}

PKI_INIT_RUNTIME_DIR=${XDG_RUNTIME_DIR}/${PKI_INIT_DIR}

# debug output:
echo XDG_RUNTIME_DIR $XDG_RUNTIME_DIR
echo PKI_INIT_RUNTIME_DIR $PKI_INIT_RUNTIME_DIR

# configuration for TLS materials:
PKI_CONFIG_JSON_DIR="res"
PKI_SETUP_VAULT_FILE=${PKI_CONFIG_JSON_DIR}"/pkisetup-vault.json"
PKI_SETUP_KONG_FILE=${PKI_CONFIG_JSON_DIR}"/pkisetup-kong.json"

# check if files exists
if [ ! -f "${PKI_SETUP_VAULT_FILE}" ]; then
    echo "Error: certificate config file for Vault is missing"
    exit 1
fi

if [ ! -f "${PKI_SETUP_KONG_FILE}" ]; then
    echo "Error: certificate config file for Kong is missing"
    exit 1
fi

# the working dir should be in vault dir based upon the current Docker image
BASE_DIR="${BASE_DIR:-/vault}"
cd $BASE_DIR
CERT_DIR=$(jq -r '.working_dir' ${PKI_SETUP_VAULT_FILE})
CERT_SUBDIR=$(jq -r '.pki_setup_dir' ${PKI_SETUP_VAULT_FILE})
ROOT_NAME=$(jq -r '.x509_root_ca_parameters | .ca_name' ${PKI_SETUP_VAULT_FILE})
CERT_EXEC="${CERT_EXEC:-./security-secrets-setup}"
# check to see if the root certificate generate with security-secrets-setup already exists
# if so then do not generate a new set of them
if [ ! -f "$CERT_DIR/$CERT_SUBDIR/$ROOT_NAME/$ROOT_NAME.pem" ]; then
    ${CERT_EXEC} --config ${PKI_SETUP_VAULT_FILE}
    [ $? -eq 0 ] || (echo "failed to generate TLS assets for Vault" && exit 1)

    ${CERT_EXEC} --config ${PKI_SETUP_KONG_FILE}
    [ $? -eq 0 ] || (echo "failed to generate TLS assets for Kong" && exit 1)

    # delete CA private key
    rm "$PWD/$CERT_DIR/$CERT_SUBDIR/$ROOT_NAME/$ROOT_NAME.priv.key"
    [ $? -eq 0 ] || (echo "failed to delete sensitive CA private key" && exit 1)

    # take ownership
    chown -R vault:vault $PWD/$CERT_DIR/$CERT_SUBDIR || true
fi 

# run the Vault's docker-entry script 
source /usr/local/bin/docker-entrypoint.sh
