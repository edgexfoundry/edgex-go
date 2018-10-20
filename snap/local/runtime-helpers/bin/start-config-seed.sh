#!/bin/bash

# wait for consul to come up
$SNAP/bin/wait-for-consul.sh "config-seed"

cd "$SNAP_DATA"/config/config-seed

"$SNAP"/bin/config-seed -c "$SNAP_DATA"/config

systemd-notify --ready
