#!/bin/bash
#
# Copyright (c) 2018
# Mainflux
#
# SPDX-License-Identifier: Apache-2.0
#

###
# Launches all EdgeX Go binaries (must be previously built).
#
# Expects that Consul and MongoDB are already installed and running.
#
###

DIR=$PWD
SERVICES=(config-seed export-client core-metadata core-command support-logging \
  support-notifications sys-mgmt-executor sys-mgmt-agent support-scheduler \
  core-data export-distro)

# Kill all edgex-* stuff
function cleanup {
	pkill edgex
}

function run {
  cd ../cmd/$1
  exec -a edgex-$1 ./$1 &
  cd $DIR
}

for i in "${SERVICES[@]}"; do
  run $i
done

trap cleanup EXIT

while : ; do sleep 1 ; done
