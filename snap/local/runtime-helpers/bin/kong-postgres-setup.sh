#!/bin/sh

# this script is called from the install hook to set up postgres
# and also from the post-refresh hook. It handles these three scenerios:
# - new install: create config file and database
# - refresh from 2.0/stable or older - a snap using postgres 10
# - refresh from a 2.1/stable or newer snap using postgres 12

# postgres configuration
CONFIG_POSTGRES_PASSWORD_DIR="$SNAP_DATA/config/postgres"
CONFIG_POSTGRES_PASSWORD_FILE="$SNAP_DATA/config/postgres/kongpw"

CONFIG_POSTGRES_12_CONFIGFILE_DIR="$SNAP_DATA/etc/postgresql/12/main"
CONFIG_POSTGRES_10_CONFIGFILE_DIR="$SNAP_DATA/etc/postgresql/10/main"
CONFIG_POSTGRES_CONFIGFILENAME="postgresql.conf"
# postgres directory
POSTGRES_DIR="$SNAP_DATA/postgresql"

# postgres 12 directories
POSTGRES_12_DIR="$SNAP_DATA/postgresql/12"
POSTGRES_12_DB_DIR="$SNAP_DATA/postgresql/12/main"

# postgres 10 directories
POSTGRES_10_DIR="$SNAP_DATA/postgresql/10"
POSTGRES_10_DB_DIR="$SNAP_DATA/postgresql/10/main"

snap_daemon_run_command() {
    export LC_ALL="C.UTF-8"
    export LANG="C.UTF-8"
    export PGHOST="$SNAP_COMMON/sockets"
    export SNAPCRAFT_PRELOAD_REDIRECT_ONLY_SHM="1"
    
    # "$SNAP/bin/snapcraft-preload" sets up SNAPCRAFT_PRELOAD and LD_PRELOAD
    #"$SNAP/snap/command-chain/snapcraft-runner" sets up PATH and LD_LIBRARY_PATH
    "$SNAP/bin/snapcraft-preload" "$SNAP/snap/command-chain/snapcraft-runner" "$SNAP/usr/bin/setpriv" --clear-groups --reuid snap_daemon --regid snap_daemon -- "$@"
}

postgres_run_command() { 
    snap_daemon_run_command "$SNAP/bin/perl5lib-launch.sh" "$@"  
}



postgres_create_password_file() {

    logger "edgex-kong-postgres-setup: creating kong postgres password" 
    # generate a random password using the automatic password generator (apg)
    # debian package sourced from the Ubuntu 18.04 archive as a snap stage-package.
    #
    # -M ncl -- says the generator should use lowercase, uppercase, and numeric symbols
    # -n 1   -- generate a single password
    # -x 24  -- maximum password len
    # -m 16  -- minimum password len
    PGPASSWD=$("$SNAP/usr/lib/apg/apg" -a 0 -M ncl -n 1 -x 24 -m 16)
    echo "$PGPASSWD" > "$CONFIG_POSTGRES_PASSWORD_FILE"

}

postgres_read_current_password_file() {
    export PGPASSWORD=`cat $CONFIG_POSTGRES_PASSWORD_FILE`
}


 

postgres_create_kong_db() {

    logger "edgex-kong-postgres-setup: creating kong postgres database"
    # add a kong user and database in postgres - note we have to run these through
    # the perl5lib-launch scripts to setup env vars properly and we need to loop
    # trying to do this because we have to wait for postgres to start up
    iter_num=0
    MAX_POSTGRES_INIT_ITERATIONS=10

    until postgres_run_command "$SNAP/usr/bin/createdb" "kong"; do
        sleep 1  
        iter_num=$(( iter_num + 1 ))
        if [ $iter_num -gt $MAX_POSTGRES_INIT_ITERATIONS ]; then
            logger "edgex-kong-postgres-setup: failed to connect to postgres after $iter_num iterations"
            exit 1
        fi
    done
    
    logger "edgex-kong-postgres-setup: creating kong postgres role"
    # createuser doesn't support specification of a password, so use psql instead.
    # Also as psql will use the database 'snap_daemon' by default, specify 'kong'
    # via environment variable.
    export PGDATABASE="kong"
    postgres_run_command "$SNAP/usr/bin/psql" "-c CREATE ROLE kong WITH NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN PASSWORD '$PGPASSWD'"

}

postgres_create_directories() {

    logger "edgex-kong-postgres-setup: creating postgres directories" 

    # ensure the postgres data directory exists and is owned by snap_daemon
    mkdir -p "$POSTGRES_DIR"
    chown -R snap_daemon:snap_daemon "$POSTGRES_DIR"
    
    # ensure the sockets dir exists and is owned by snap_daemon
    mkdir -p "$SNAP_COMMON/sockets"
    chown -R snap_daemon:snap_daemon "$SNAP_COMMON/sockets"

    # create the directory used for the kong backup
    mkdir -p "$SNAP_COMMON/refresh"
    chown -R snap_daemon:snap_daemon "$SNAP_COMMON/refresh"

    # create the config directory used for the kong password
    mkdir -p $CONFIG_POSTGRES_PASSWORD_DIR 

    # create directory for configuration
    mkdir -p $CONFIG_POSTGRES_12_CONFIGFILE_DIR

}

postgres_init_postgres_database() {
    logger "edgex-kong-postgres-setup: initializing database (initdb -D $POSTGRES_12_DB_DIR)" 
    snap_daemon_run_command "$SNAP/usr/lib/postgresql/12/bin/initdb" -D "$POSTGRES_12_DB_DIR"
}


postgres_create_configuration_file() {

    logger "edgex-kong-postgres-setup: creating configuration file $CONFIG_POSTGRES_12_CONFIGFILE_DIR/$CONFIG_POSTGRES_CONFIGFILENAME"
    
    cp "$SNAP/config/postgres/$CONFIG_POSTGRES_CONFIGFILENAME" "$CONFIG_POSTGRES_12_CONFIGFILE_DIR/"

    # do replacement of the $SNAP, $SNAP_DATA, $SNAP_COMMON environment variables in the config files
    sed -i -e "s@\$SNAP_COMMON@$SNAP_COMMON@g" \
        -e "s@\$SNAP_DATA@$CURRENT@g" \
        -e "s@\$SNAP@$SNAP_CURRENT@g" \
        "$CONFIG_POSTGRES_12_CONFIGFILE_DIR/$CONFIG_POSTGRES_CONFIGFILENAME" 
    
}

postgres_remove_old_postgres() {
    logger "edgex-kong-postgres-setup: removing postgres 10"
    rm -rf "$CONFIG_POSTGRES_10_CONFIGFILE_DIR"
    snap_daemon_run_command rm -rf "$POSTGRES_10_DIR"
}

# called from the security-proxy-post-setup script.
# kong needs to be running for this to succeed, so we can't run this in the 
# post-refresh hook
postgres_restore_backup() {
    if [ -f "$SNAP_COMMON/refresh/kong.sql" ]; then
      logger "edgex-kong-postgres-setup: restoring kong database from $SNAP_COMMON/refresh/kong.sql"
      postgres_run_command "$SNAP/usr/bin/psql" "-Ukong" "-f$SNAP_COMMON/refresh/kong.sql" "kong"
    else
        logger "edgex-kong-postgres-setup: No backup found at $SNAP_COMMON/refresh/kong.sql"
    fi
}


# This function is called either:
# a) from install hook (via the install-setup-postgres.sh script), to set up a new PG 12 instance
# b) from the post-refresh hook, to refresh the PG instance from an older PG version
postgres_install() {

    logger "edgex-kong-postgres-setup: install postgres 12"
  
    # Note: It's fine to replace the current postgres installation.
    # Postgresql is only used for kong and if we're doing a snap refresh, then
    # we will backup and restore the Kong configuration. 

    # If there is an old Postgres v10.x installation, then just remove it
    if [ -d $CONFIG_POSTGRES_10_CONFIGFILE_DIR ]; then
        postgres_remove_old_postgres
    fi
    
    # create the required directories 
    postgres_create_directories

    # and the config file
    postgres_create_configuration_file

    postgres_create_password_file
    
    # set up the new postgres v12 database
    postgres_init_postgres_database

    # start postgres up and wait a bit for it so we can create the database and user for kong
    logger "edgex-kong-postgres-setup: install - starting postgres"
    snapctl start "$SNAP_NAME.postgres"
 
    postgres_create_kong_db

    # Load the SQL file with the kong backup from the older postgres server
    postgres_restore_backup

    # stop postgres again in case the security services should be turned off
    logger "edgex-kong-postgres-setup: install - stopping postgres"
    snapctl stop "$SNAP_NAME.postgres"

    # modify postgres authentication config to use 'md5' (password)
    snap_daemon_run_command sed -i -e "s@trust@md5@g" "$POSTGRES_12_DB_DIR/pg_hba.conf"
}


# called from the pre-refresh hook. Used to backup the Kong configuration
kong_pre_refresh() {
    logger "edgex-kong-postgres-setup: pre_refresh dumping kong database to $SNAP_COMMON/refresh/kong.sql"
    postgres_read_current_password_file
    postgres_run_command "$SNAP/usr/bin/pg_dump" "-Ukong" "-f$SNAP_COMMON/refresh/kong.sql" "kong"
}

# The post-refresh hook is called in the context of the new snap
# revision, and prior to services being started.
# It calls this function, which will upgrade postgres if required and
# migrate the kong database
kong_post_refresh() {
    logger "edgex-kong-postgres-setup: post-refresh"

    # Force remove kong-admin-jwt due to the fact that it's
    # written by security-secretstore-setup (2.0.0-*) with
    # read-only perm for the owner.
    # See: https://github.com/edgexfoundry/edgex-go/issues/3818
    if [ -f "$SNAP_DATA/secrets/security-proxy-setup/kong-admin-jwt" ]; then
        logger "edgex-kong-postgres-setup: removing stale kong-admin-jwt"
        rm -f "$SNAP_DATA/secrets/security-proxy-setup/kong-admin-jwt"
    fi

    if [ -d $CONFIG_POSTGRES_10_CONFIGFILE_DIR ]; then 
      postgres_install
    else
     logger "edgex-kong-postgres-setup: post-refresh - nothing to refresh"
    fi

    snap_daemon_run_command rm -f $SNAP_COMMON/refresh/kong.sql
}


method=${1:-"post-stop-command"}
logger "edgex-kong-postgres-setup: calling $method"

case $method in 

post-refresh)
    kong_post_refresh
    ;;
pre-refresh)
    kong_pre_refresh
    ;;
install)
    postgres_install
    ;;
esac