#!/bin/sh

# this script is called from the install hook to set up postgres
# and also from the post-refresh hook. It handles these three scenerios:
# - new install: create config file and database
# - refresh from a snap using postgres 10
# - refresh from a newer snap using postgres 12



# postgres configuration
KONG_CONFIG_PASSWORD_DIR="$SNAP_DATA/config/postgres"
KONG_CONFIG_PASSWORD_FILE="$SNAP_DATA/config/postgres/kongpw"

POSTGRES_10_ETC_CONFIG_DIR="$SNAP_DATA/etc/postgresql/10/main"

POSTGRES_CONFIG_FILENAME="postgresql.conf"

POSTGRES_12_DB_DIR="$SNAP_DATA/postgresql/12/main" 

# if we use the postgres binaries in /bin, then --devmode install fails.
PSQL_BIN_PATH="$SNAP/usr/lib/postgresql/12/bin/"

snap_setup_environment() {
    logger "edgex-kong-postgres-setup: setting up environment" 
    export LC_ALL="C.UTF-8"
    export LANG="C.UTF-8"
    export PGHOST="$SNAP_COMMON/sockets"
    export SNAPCRAFT_PRELOAD_REDIRECT_ONLY_SHM="1"
    
    # "$SNAP/bin/snapcraft-preload" sets up SNAPCRAFT_PRELOAD and LD_PRELOAD
    # export SNAPCRAFT_PRELOAD=$SNAP
    # export LD_PRELOAD="$SNAP/lib/libsnapcraft-preload.so"
    . "$SNAP/bin/snapcraft-preload"
    
    #"$SNAP/snap/command-chain/snapcraft-runner" sets up PATH and LD_LIBRARY_PATH
    # export PATH=...
    # export LD_LIBRARY_PATH=...
    . "$SNAP/snap/command-chain/snapcraft-runner"

    # setup perl5
    . "$SNAP/bin/perl5lib-launch.sh"
}

snap_daemon_run_command() {
     "$SNAP/usr/bin/setpriv" --clear-groups --reuid snap_daemon --regid snap_daemon -- "$@"
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
    echo "$PGPASSWD" > "$KONG_CONFIG_PASSWORD_FILE"

}

 


 postgres_init_postgres_database() {
    logger "edgex-kong-postgres-setup: initializing database (initdb -D $POSTGRES_12_DB_DIR)" 
    snap_daemon_run_command "$PSQL_BIN_PATH/initdb" -D "$POSTGRES_12_DB_DIR"
}


postgres_create_kong_db() {

    logger "edgex-kong-postgres-setup: creating kong postgres database"
    # add a kong user and database in postgres - note we have to run these through
    # the perl5lib-launch scripts to setup env vars properly and we need to loop
    # trying to do this because we have to wait for postgres to start up
    iter_num=0
    MAX_POSTGRES_INIT_ITERATIONS=10

    until snap_daemon_run_command "$PSQL_BIN_PATH/createdb" "kong"; do
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
    
   until snap_daemon_run_command "$PSQL_BIN_PATH/psql" -c "CREATE ROLE kong WITH NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN PASSWORD '$PGPASSWD'"; do
        sleep 1  
        iter_num=$(( iter_num + 1 ))
        if [ $iter_num -gt $MAX_POSTGRES_INIT_ITERATIONS ]; then
            logger "edgex-kong-postgres-setup: failed to create kong role  after $iter_num iterations"
            exit 1
        fi
    done

}

postgres_setup_directories() {

    logger "edgex-kong-postgres-setup: creating postgres directories" 

    # ensure the postgres data directory exists and is owned by snap_daemon
    mkdir -p "$POSTGRES_12_DB_DIR"
    chown -R snap_daemon:snap_daemon "$SNAP_DATA/postgresql" 
    
    # ensure the sockets dir exists and is owned by snap_daemon
    mkdir -p "$SNAP_COMMON/sockets"
    chown -R snap_daemon:snap_daemon "$SNAP_COMMON/sockets"

    # create the directory used for the kong backup
    mkdir -p "$SNAP_COMMON/refresh"
    chown -R snap_daemon:snap_daemon "$SNAP_COMMON/refresh"

    # create the config directory used for the kong password
    mkdir -p $KONG_CONFIG_PASSWORD_DIR 

    # remove pg 10 directories
    rm -rf "$SNAP_DATA/etc/postgresql"
    
    # remove database directory
    rm -rf "$SNAP_DATA/postgresql/10/main"
}



postgres_create_configuration_file() {

    logger "edgex-kong-postgres-setup: creating configuration file $POSTGRES_12_DB_DIR/$POSTGRES_CONFIG_FILENAME"
    
   snap_daemon_run_command cp "$SNAP/config/postgres/$POSTGRES_CONFIG_FILENAME" "$POSTGRES_12_DB_DIR/"

    # do replacement of the $SNAP, $SNAP_DATA, $SNAP_COMMON environment variables in the config files
     snap_daemon_run_command sed -i -e "s@\$SNAP_COMMON@$SNAP_COMMON@g" \
        -e "s@\$SNAP_DATA@$CURRENT@g" \
        -e "s@\$SNAP@$SNAP_CURRENT@g" \
        "$POSTGRES_12_DB_DIR/$POSTGRES_CONFIG_FILENAME" 
    
}
 

# called from the security-proxy-post-setup script.
# kong needs to be running for this to succeed, so we can't run this in the 
# post-refresh hook
postgres_restore_backup() {
    if [ -f "$SNAP_COMMON/refresh/kong.sql" ]; then
      logger "edgex-kong-postgres-setup: restoring kong database from $SNAP_COMMON/refresh/kong.sql"
      snap_daemon_run_command "$PSQL_BIN_PATH/psql" "-Ukong" "-f$SNAP_COMMON/refresh/kong.sql" "kong"
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

    # create the required directories 
    postgres_setup_directories

    postgres_create_password_file
    
    # set up the new postgres v12 database
    postgres_init_postgres_database

    # and the config file
    postgres_create_configuration_file

    # start postgres and create the database and user for kong
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
    
    # use the kong role password
    export PGPASSWORD=`cat $KONG_CONFIG_PASSWORD_FILE`

    snap_daemon_run_command "$SNAP/usr/bin/pg_dump" "-Ukong" "-f$SNAP_COMMON/refresh/kong.sql" "kong"
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

    if [ -d $POSTGRES_10_ETC_CONFIG_DIR ]; then
      postgres_install
    else
     logger "edgex-kong-postgres-setup: post-refresh - nothing to refresh"
    fi

    # remove the SQL file
    snap_daemon_run_command rm -f $SNAP_COMMON/refresh/kong.sql
}


method=${1:-"post-stop-command"}
logger "edgex-kong-postgres-setup: calling $method"

snap_setup_environment

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