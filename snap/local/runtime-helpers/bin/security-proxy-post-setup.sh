#!/bin/bash -e

# This script is called as a post-stop-command when
# security-proxy-setup oneshot service stops.

logger "edgexfoundry:security-proxy-post-setup"

# Add bin directory to have e.g. secrets-config in PATH
export PATH="$SNAP/bin:$PATH"

# Process snap options that rely on the started Security Proxy
$SNAP/bin/helper-go options --app=secrets-config
