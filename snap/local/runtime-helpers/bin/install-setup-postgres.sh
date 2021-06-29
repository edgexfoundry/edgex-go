#!/bin/sh

# setup postgres db config file with env vars replaced
if [ ! -f "$SNAP_DATA/etc/postgresql/10/main/postgresql.conf" ]; then
    mkdir -p "$SNAP_DATA/etc/postgresql/10/main"
    cp "$SNAP/etc/postgresql/10/main/postgresql.conf" "$SNAP_DATA/etc/postgresql/10/main/postgresql.conf"
    # do replacement of the $SNAP, $SNAP_DATA, $SNAP_COMMON environment variables in the config files
    sed -i -e "s@\$SNAP_COMMON@$SNAP_COMMON@g" \
        -e "s@\$SNAP_DATA@$SNAP_DATA_CURRENT@g" \
        -e "s@\$SNAP@$SNAP_CURRENT@g" \
        "$SNAP_DATA/etc/postgresql/10/main/postgresql.conf"
fi

# ensure the postgres data directory exists and is owned by snap_daemon
mkdir -p "$SNAP_DATA/postgresql"
chown -R snap_daemon:snap_daemon "$SNAP_DATA/postgresql"

# setup the postgres data directory
"$SNAP/bin/drop-snap-daemon.sh" "$SNAP/usr/lib/postgresql/10/bin/initdb" -D "$SNAP_DATA/postgresql/10/main"

# ensure the sockets dir exists and is properly owned
mkdir -p "$SNAP_COMMON/sockets"
chown -R snap_daemon:snap_daemon "$SNAP_COMMON/sockets"

# start postgres up and wait a bit for it so we can create the database and user
# for kong
snapctl start "$SNAP_NAME.postgres"

# add a kong user and database in postgres - note we have to run these through
# the perl5lib-launch scripts to setup env vars properly and we need to loop
# trying to do this because we have to wait for postgres to start up
iter_num=0
MAX_POSTGRES_INIT_ITERATIONS=10
until "$SNAP/bin/drop-snap-daemon.sh" "$SNAP/bin/perl5lib-launch.sh" "$SNAP/usr/bin/createdb" kong; do
    sleep 1
    iter_num=$(( iter_num + 1 ))
    if [ $iter_num -gt $MAX_POSTGRES_INIT_ITERATIONS ]; then
        logger "edgexfoundry:install: failed to create kong db in postgres after $iter_num iterations"
        exit 1
    fi
done

# generate a random password using the automatic password generator (apg)
# debian package sourced from the Ubuntu 18.04 archive as a snap stage-package.
#
# -M ncl -- says the generator should use lowercase, uppercase, and numeric symbols
# -n 1   -- generate a single password
# -x 24  -- maximum password len
# -m 16  -- minimum password len
PGPASSWD=$("$SNAP/usr/lib/apg/apg" -a 0 -M ncl -n 1 -x 24 -m 16)
mkdir -p "$SNAP_DATA/config/postgres/"
echo "$PGPASSWD" > "$SNAP_DATA/config/postgres/kongpw"

# createuser doesn't support specification of a password, so use psql instead.
# Also as psql will use the database 'snap_daemon' by default, specify 'kong'
# via environment variable.
export PGDATABASE="kong"
iter_num=0
until "$SNAP/bin/drop-snap-daemon.sh" "$SNAP/bin/perl5lib-launch.sh" "$SNAP/usr/bin/psql" \
    "-c CREATE ROLE kong WITH NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN PASSWORD '$PGPASSWD'"; do
    sleep 1
    iter_num=$(( iter_num + 1 ))
    if [ $iter_num -gt $MAX_POSTGRES_INIT_ITERATIONS ]; then
        logger "edgexfoundry:install: failed to create kong user in postgres after $iter_num iterations"
        exit 1
    fi
done

# stop postgres again in case the security services should be turned off
snapctl stop "$SNAP_NAME.postgres"

# modify postgres authentication config to use 'md5' (password)
"$SNAP/bin/drop-snap-daemon.sh" sed -i -e "s@trust@md5@g" "$SNAP_DATA/postgresql/10/main/pg_hba.conf"
