#!/bin/bash -e

# This script is called as a post-stop-command when
# security-proxy-setup oneshot service stops.

logger "edgexfoundry:security-proxy-post-setup"

# add bin directory to have e.g. secret-config in PATH
export PATH="$SNAP/bin:$PATH"

# Several config options depend on resources that only exist after proxy 
# setup. This re-applies the config options logic after deferred startup:
$SNAP/snap/hooks/configure options --service=security-proxy
