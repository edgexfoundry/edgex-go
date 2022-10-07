#!/bin/bash -e

# This script is called as a post-stop-command when
# security-proxy-setup oneshot service stops.

logger "edgexfoundry:security-proxy-post-setup"

# add bin directory to have e.g. secret-config in PATH
export PATH="$SNAP/bin:$PATH"

# Several config options depend on resources that only exist after proxy 
# setup. This re-applies the config options logic after deferred startup:
$SNAP/bin/helper-go options --service=security-proxy

# Process the EdgeX >=2.2 snap options
$SNAP/bin/helper-go options --service=secrets-config
