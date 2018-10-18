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
cd $CMD/support-logging
# Add `edgex-` prefix on start, so we can find the process family
exec -a edgex-support-logging ./support-logging &
cd $DIR

###
# Core Command
###
cd $CMD/core-command
# Add `edgex-` prefix on start, so we can find the process family
exec -a edgex-core-command ./core-command &
cd $DIR

###
# Core Data
###
cd $CMD/core-data
exec -a edgex-core-data ./core-data &
cd $DIR

###
# Core Metadata
###
cd $CMD/core-metadata
exec -a edgex-core-metadata ./core-metadata &
cd $DIR

###
# Export Client
###
cd $CMD/export-client
exec -a edgex-export-client ./export-client &
cd $DIR

###
# Export Distro
###
cd $CMD/export-distro
exec -a edgex-export-distro ./export-distro &
cd $DIR

###
# Support Notifications
###
cd $CMD/support-notifications
# Add `edgex-` prefix on start, so we can find the process family
exec -a edgex-support-notifications ./support-notifications &
cd $DIR

###
# System Management Agent
###
cd $CMD/sys-mgmt-agent
# Add `edgex-` prefix on start, so we can find the process family
exec -a edgex-sys-mgmt-agent ./sys-mgmt-agent &
cd $DIR

# Support Scheduler
###
cd $CMD/support-scheduler
# Add `edgex-` prefix on start, so we can find the process family
exec -a edgex-support-scheduler ./support-scheduler &
cd $DIR

trap cleanup EXIT

while : ; do sleep 1 ; done