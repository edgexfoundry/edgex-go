#!/bin/bash -e

# this is a workaround to prevent kong from running on arm64 until kong 
# upstream supports running on arm64 properly, see 
# https://github.com/edgexfoundry/blackbox-testing/issues/185 for more details
# also note that we disable kong from the install hook, but that is only
# valid on first install, any refreshes will trigger it to be restarted due to
# https://bugs.launchpad.net/snapd/+bug/1818306 , hence this workaround
if [ "$SNAP_ARCH" = "arm64" ]; then
  exit 0
fi

# the kong wrapper script from $SNAP
export KONG_SNAP="$SNAP/bin/kong-wrapper.sh"

# run kong migrations up to bootstrap the cassandra database
# note that sometimes cassandra can be in a "starting up" state, etc.
# and in this case we should just loop and keep trying
# we don't implement a timeout here because systemd will kill us if we 
# don't succeed in 15 minutes (or whatever the configured stop-timeout is)
until $KONG_SNAP migrations bootstrap --conf "$KONG_CONF"; do
    sleep 5
done

# now start kong normally
$KONG_SNAP start --conf "$KONG_CONF"
