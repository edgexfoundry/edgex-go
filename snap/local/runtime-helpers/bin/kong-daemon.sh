#!/bin/bash -e

KONG=/usr/local/bin/kong

# run kong migrations up to bootstrap the database
# note that sometimes the database can be in a "starting up" state, etc.
# and in this case we should just loop and keep trying
# we don't implement a timeout here because systemd will kill us if we 
# don't succeed in 15 minutes (or whatever the configured stop-timeout is)
until "$KONG" migrations up --conf "$KONG_CONF"; do
    sleep 5
done

# set up Kong's admin API plugins via kong.yml:
"$KONG" config db_import "$KONGADMIN_CONFIGFILEPATH" || true

# now start kong normally
"$KONG" start --conf "$KONG_CONF" --v
