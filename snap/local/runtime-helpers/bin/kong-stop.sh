#!/bin/bash

# stop kong
/usr/local/bin/kong stop -p "$SNAP_DATA/kong"

# in some cases stopping kong doesn't succeed properly, so to ensure that
# it always is put into a state where it can startup, just remove the env
# file in case it somehow still exists, then the next invocation of kong
# will always be able to start
rm -f "$SNAP_DATA/kong/.kong_env"
