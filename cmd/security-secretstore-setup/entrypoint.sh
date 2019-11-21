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

echo "creating /vault/config/assets"

# create token directory and
# grant permissions of folders for vault:vault
mkdir -p /vault/config/assets
chown -Rh 100:1000 /vault/

echo "starting vault-worker..."

# need to wait until security-secrets-setup produces the necessary TLS assets
# before start the vault-worker

until /consul/scripts/consul-svc-healthy.sh security-secrets-setup; do 
    echo 'waiting for security-secrets-setup'; 
    sleep 1; 
done;

/security-secretstore-setup --init=true --vaultInterval=10

exec "$@"
