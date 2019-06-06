#!/bin/bash -e

# push the configuration files into consul
"$SNAP/bin/config-seed" \
    --cmd "$SNAP_DATA/config" \
    -confdir "$SNAP_DATA/config/config-seed/res" \
    --props "$SNAP_DATA/config/config-seed/res/properties" \
    -overwrite

# if no arguments were provided, then restart all services that are currently
# running
if [ $# -eq 0 ]; then
    # restart all active edgex services to ensure that they pick up their new 
    # configuration from consul
    # for now, limit outselves to the core-*, export-*, support-*, device-*,
    # sys-mgmt-agent, and security-service helper services
    # this means if a user changes i.e. kong configuration they will need to
    # restart kong-daemon manually
    # TODO: maybe implement some kind of file hashing to determine which services
    # had their configs changed and only restart changed services?
    for svc in  $(snapctl services | grep "core-*\|export-*\|support-*\|sys-mgmt-agent\|device-*\|vault-worker\|edgexproxy" | grep -v inactive | grep active | awk '{print $1}');  do
        snapctl restart "$svc"
    done
fi

# otherwise restart the args provided, assuming they are all names of
# services in the snap
set +e 
for svc in "$@"; do
    # check if it's a known service - if not fail
    SNAP_NAME_SVC="$SNAP_NAME.$svc"
    if ! snapctl services | grep -q "$SNAP_NAME_SVC" ; then
        echo "unknown service \"$svc\""
        exit 1
    fi
    # check if it's running - if so restart
    if  snapctl services | grep "$SNAP_NAME_SVC" | grep -q -v inactive; then
        snapctl restart "$SNAP_NAME_SVC"
    fi
done
