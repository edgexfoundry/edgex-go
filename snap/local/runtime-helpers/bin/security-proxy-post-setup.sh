#!/bin/bash -e

# This script is called as a post-stop-command when
# security-proxy-setup oneshot service stops.
#
# Several config options depend on resources that only exist after proxy is 
# setup. This scripts re-runs the configure hook after the deferred startup 
# of security-proxy-setup to apply such configurations.

logger "edgex-secretstore-proxy:post-setup: calling configure hook"

# add bin directory to have e.g. secret-config in PATH
export PATH="$SNAP/bin:$PATH"

$SNAP/snap/hooks/configure
