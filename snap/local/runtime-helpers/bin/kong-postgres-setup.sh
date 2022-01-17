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
    "$SNAP/usr/bin/setpriv" --clear-groups --reuid snap_daemon --regid snap_daemon -- "$@"
}

postgres_run_command() { 
   snap_daemon_run_command "$SNAP/bin/perl5lib-launch.sh" "$@"  

}



postgres_create_password_file() {

    logger "edgexfoundry:postgres: postgres_create_password_file" 
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

    logger "edgexfoundry:postgres: postgres_create_kong_db"
    # add a kong user and database in postgres - note we have to run these through
    # the perl5lib-launch scripts to setup env vars properly and we need to loop
    # trying to do this because we have to wait for postgres to start up
    iter_num=0
    MAX_POSTGRES_INIT_ITERATIONS=10

    until postgres_run_command "$SNAP/usr/bin/createdb" "kong"; do
        logger "edgexfoundry:postgres: Created kong db in postgres" 
        sleep 1  
        iter_num=$(( iter_num + 1 ))
        if [ $iter_num -gt $MAX_POSTGRES_INIT_ITERATIONS ]; then
            logger "edgexfoundry:postgres: failed to connect to postgres after $iter_num iterations"
            exit 1
        fi
    done
    
    logger "edgexfoundry:postgres: postgres_create_kong_user"

    # createuser doesn't support specification of a password, so use psql instead.
    # Also as psql will use the database 'snap_daemon' by default, specify 'kong'
    # via environment variable.
    export PGDATABASE="kong"
    postgres_run_command "$SNAP/usr/bin/psql" "-c CREATE ROLE kong WITH NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN PASSWORD '$PGPASSWD'"
    logger "edgexfoundry:postgres: Created kong user"   

}

postgres_create_directories() {

    logger "edgexfoundry:postgres: postgres_create_directories" 

    # ensure the postgres data directory exists and is owned by snap_daemon
    logger "edgexfoundry:postgres: creating $POSTGRES_DIR"
    mkdir -p "$POSTGRES_DIR"
    chown -R snap_daemon:snap_daemon "$POSTGRES_DIR"
    
    # ensure the sockets dir exists and is owned by snap_daemon
    logger "edgexfoundry:postgres: creating $SNAP_COMMON/sockets"
    mkdir -p "$SNAP_COMMON/sockets"
    chown -R snap_daemon:snap_daemon "$SNAP_COMMON/sockets"

    mkdir -p "$SNAP_COMMON/refresh"
    chown -R snap_daemon:snap_daemon "$SNAP_COMMON/refresh"


    # create the config directory used for the kong password
    logger "edgexfoundry:postgres: creating $CONFIG_POSTGRES_PASSWORD_DIR"
    mkdir -p $CONFIG_POSTGRES_PASSWORD_DIR 

    # create directory for configuration
    mkdir -p $CONFIG_POSTGRES_12_CONFIGFILE_DIR

}

postgres_create_postgres_database() {
    logger "edgexfoundry:postgres: postgres_create_postgres_database" 
    logger "edgexfoundry:postgres: initdb -D $POSTGRES_12_DB_DIR"  
    snap_daemon_run_command "$SNAP/usr/lib/postgresql/12/bin/initdb" -D "$POSTGRES_12_DB_DIR"
}


postgres_create_configuration_file() {

    logger "edgexfoundry:postgres: postgres_create_configuration"
    logger "edgexfoundry:postgres: setting up new configuration file at $CONFIG_POSTGRES_12_CONFIGFILE_DIR/$CONFIG_POSTGRES_CONFIGFILENAME"
    
    cp "$SNAP/config/postgres/$CONFIG_POSTGRES_CONFIGFILENAME" "$CONFIG_POSTGRES_12_CONFIGFILE_DIR/"

    # do replacement of the $SNAP, $SNAP_DATA, $SNAP_COMMON environment variables in the config files
    sed -i -e "s@\$SNAP_COMMON@$SNAP_COMMON@g" \
        -e "s@\$SNAP_DATA@$CURRENT@g" \
        -e "s@\$SNAP@$SNAP_CURRENT@g" \
        "$CONFIG_POSTGRES_12_CONFIGFILE_DIR/$CONFIG_POSTGRES_CONFIGFILENAME" 
    
}

postgres_remove_old_postgres() {
    logger "edgexfoundry:postgres: postgres_remove_old_postgres"
    rm -rf "$CONFIG_POSTGRES_10_CONFIGFILE_DIR"
    snap_daemon_run_command rm -rf "$POSTGRES_10_DIR"
}

# called from the security-proxy-post-setup script.
# kong needs to be running for this to succeed, so we can't run this in the 
# post-refresh hook
kong_restore_backup() {
    if [ -f "$SNAP_COMMON/refresh/kong.sql" ]; then
      logger "edgexfoundry:postgres: refresh - loading saved kong data"
      postgres_run_command "$SNAP/usr/bin/psql" "-Ukong" "-f$SNAP_COMMON/refresh/kong.sql" "kong"
    fi
}

postgres_install() {
     logger "edgexfoundry:postgres: postgres_install"
 
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
    postgres_create_postgres_database

    # start postgres up and wait a bit for it so we can create the database and user for kong
    logger "edgexfoundry:postgres: refresh - starting postgres"
    snapctl start "$SNAP_NAME.postgres"
 
    postgres_create_kong_db

    # Load the SQL file with the kong backup from the older postgres server
    kong_restore_backup


    # stop postgres again in case the security services should be turned off
    logger "edgexfoundry:postgres: refresh - stopping postgres"
    snapctl stop "$SNAP_NAME.postgres"

    # modify postgres authentication config to use 'md5' (password)
    snap_daemon_run_command sed -i -e "s@trust@md5@g" "$POSTGRES_12_DB_DIR/pg_hba.conf"
}


# called from the pre-refresh hook. Used to backup the Kong configuration
kong_pre_refresh() {
    logger "edgexfoundry:postgres: postgres_pre_refresh"
    postgres_read_current_password_file
    postgres_run_command "$SNAP/usr/bin/pg_dump" "-Ukong" "-f$SNAP_COMMON/refresh/kong.sql" "kong"
}

# called from the post-refresh hook to set up a new postgres 12 deployment
# if needed and restore the kong configuration
kong_post_refresh() {
    logger "edgexfoundry:postgres: postgres_post_refresh"
    if [ -d $CONFIG_POSTGRES_10_CONFIGFILE_DIR ]; then
      postgres_install
    else
     logger "edgexfoundry:postgres: post-refresh - nothing to refresh"
    fi
}





