#!/bin/sh -e

#
# Copyright (c) 2019 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software distributed under the License
# is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
# or implied. See the License for the specific language governing permissions and limitations under
# the License.
#
# SPDX-License-Identifier: Apache-2.0'
#

make -C ../../.. cmd/security-file-token-provider/security-file-token-provider

cp ../res/token-config.json res/

docker-compose -f docker-compose.yml up -d

echo 'Waiting for stack to come up...'
sleep 5

vaultid=`docker ps | grep 'edgexfoundry/docker-edgex-vault' | cut -d' ' -f1`
echo "Vault container id $vaultid"

docker exec $vaultid cat /vault/config/pki/EdgeXFoundryCA/EdgeXFoundryCA.pem > EdgeXFoundryCA.pem
docker exec $vaultid cat /vault/config/assets/resp-init.json > resp-init.json

vault_token=`jq -r .root_token ./resp-init.json`
echo "Vault root token is $vault_token"

cat <<EOH > res/configuration.toml
[SecretService]
Scheme = "https"
Server = "localhost"
Port = 8200
# CaFilePath = "EdgeXFoundryCA.pem"

[TokenFileProvider]
PrivilegedTokenPath = "resp-init.json"
ConfigFile = "res/token-config.json"
OutputDir = "/tmp"
OutputFilename = "secrets-token.json"

[Writable]
LogLevel = 'DEBUG'

[Logging]
EnableRemote = false
File = './logs/security-file-token-provider.log'
EOH

echo "---==[ BEFORE policies ]==---"
docker exec -e "VAULT_ADDR=https://edgex-vault:8200" -e "VAULT_TOKEN=$vault_token" $vaultid vault policy list -tls-skip-verify

echo "---==[ BEFORE tokens ]==---"
docker exec -e "VAULT_ADDR=https://edgex-vault:8200" -e "VAULT_TOKEN=$vault_token" $vaultid vault list -tls-skip-verify auth/token/accessors

../security-file-token-provider

echo "---==[ AFTER policies ]==---"
docker exec -e "VAULT_ADDR=https://edgex-vault:8200" -e "VAULT_TOKEN=$vault_token" $vaultid vault policy list -tls-skip-verify

echo "---==[ AFTER policy (dump edgex-service-service-name) ]==---"
docker exec -e "VAULT_ADDR=https://edgex-vault:8200" -e "VAULT_TOKEN=$vault_token" $vaultid vault policy read -tls-skip-verify edgex-service-service-name

echo "---==[ AFTER tokens ]==---"
docker exec -e "VAULT_ADDR=https://edgex-vault:8200" -e "VAULT_TOKEN=$vault_token" $vaultid vault list -tls-skip-verify auth/token/accessors

new_token=`jq -r .auth.client_token /tmp/service-name/secrets-token.json`

echo "---==[ INFO on new token ]==---"
docker exec -e "VAULT_ADDR=https://edgex-vault:8200" -e "VAULT_TOKEN=$new_token" $vaultid vault token lookup -tls-skip-verify

