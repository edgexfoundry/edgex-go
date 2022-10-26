#!/bin/bash -e

KONG=/usr/local/bin/kong

# bootstrap the database
# note that sometimes the database can be in a "starting up" state, etc.
# and in this case we should just loop and keep trying
# we don't implement a timeout here because systemd will kill us if we 
# don't succeed in 15 minutes (or whatever the configured stop-timeout is)
until "$KONG" migrations bootstrap --conf "$KONG_CONF"; do
    sleep 5
done

# perform migration for upgrades
"$KONG" migrations up --conf "$KONG_CONF"

# set up Kong's admin API plugins via kong.yml:
"$KONG" config db_import "$KONGADMIN_CONFIGFILEPATH" || true

# now start kong normally
"$KONG" start --conf "$KONG_CONF" --v
