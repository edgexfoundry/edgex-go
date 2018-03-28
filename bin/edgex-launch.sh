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
CMD=../cmd

# Kill all edgex-* stuff
function cleanup {
	pkill edgex
}

###
# Support logging
###
printf "\n### Starting edgex-support-logging\n"
cd $CMD/support-logging
# Add `edgex-` prefix on start, so we can find the process family
exec -a edgex-support-logging ./support-logging &
cd $DIR

###
# Core Command
###
printf "\n### Starting edgex-core-command\n"
cd $CMD/core-command
# Add `edgex-` prefix on start, so we can find the process family
exec -a edgex-core-command ./core-command &
cd $DIR

###
# Core Data
###
printf "\n### Starting edgex-core-data\n"
cd $CMD/core-data
exec -a edgex-core-data ./core-data &
cd $DIR

###
# Core Metadata
###
printf "\n### Starting edgex-core-metadata\n"
cd $CMD/core-metadata
exec -a edgex-core-metadata ./core-metadata &
cd $DIR

###
# Export Client
###
printf "\n### Starting edgex-export-client\n"
cd $CMD/export-client
exec -a edgex-export-client ./export-client &
cd $DIR

###
# Export Distro
###
printf "\n### Starting edgex-export-distro\n"
cd $CMD/export-distro
exec -a edgex-export-distro ./export-distro &
cd $DIR


trap cleanup EXIT

while : ; do sleep 1 ; done
