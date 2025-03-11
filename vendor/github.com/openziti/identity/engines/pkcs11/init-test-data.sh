#!/bin/bash

#	Copyright 2019 NetFoundry, Inc.
#
#	Licensed under the Apache License, Version 2.0 (the "License");
#	you may not use this file except in compliance with the License.
#	You may obtain a copy of the License at
#
#	https://www.apache.org/licenses/LICENSE-2.0
#
#	Unless required by applicable law or agreed to in writing, software
#	distributed under the License is distributed on an "AS IS" BASIS,
#	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#	See the License for the specific language governing permissions and
#	limitations under the License.

conf=softhsm2.conf

lib=$1
pin=$2

[ -e ${conf} ] || exit 1

tokendir=$(awk '/^directories.tokendir/{print $3;}' ${conf})

mkdir -p ${tokendir}
export SOFTHSM2_CONF=${conf}

softhsm2-util --init-token --slot 0 --label 'ziti-test-token' --so-pin ${pin} --pin ${pin}
pkcs11-tool --module ${lib} -p ${pin} -k --key-type rsa:2048 --id 01 --label ziti-rsa-key
pkcs11-tool --module ${lib} -p ${pin} -k --key-type EC:prime256v1 --id 02 --label ziti-ecdsa-key

