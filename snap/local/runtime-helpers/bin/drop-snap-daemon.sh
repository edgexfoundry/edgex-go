#!/bin/bash -e

"$SNAP/usr/bin/setpriv" --clear-groups --reuid snap_daemon --regid snap_daemon -- "$@"
