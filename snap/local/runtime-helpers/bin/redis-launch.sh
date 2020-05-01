#!/bin/sh -e

# Note - this pathname needs to be kept in sync with the
# security-secretstore-setup-launch.sh script until the
# following edgex-go bug is fixed:
#
# https://github.com/edgexfoundry/edgex-go/issues/2503
#
REDIS5_PASSWORD=`cat "$SNAP_DATA/secrets/edgex-redis/redis5-password"`

exec "$SNAP/bin/redis-server" \
     --requirepass "$REDIS5_PASSWORD" \
     --dir "$SNAP_DATA/redis" \
     --save 900 1 \
     --save 300 10 \
     --save 60 10000
