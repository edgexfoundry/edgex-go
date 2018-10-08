#!/bin/sh
set -ex

# TODO: it would be better if this func first tried SIGTERM,
# and fell back to SIGKILL, however this script is not meant
# to be long-lived, hopefully deprecated once everything
# all micro services becomes snap daemons.
kill_service() {
	echo "sending SIGKILL to $2 ($1) service"
	kill -9 "$1" || true
}

# send SIGINT to consul
int_service() {
	echo "sending SIGINT to $2 ($1) service"
	kill -INT "$1" || true
}

# Bootstrap service env vars
if [ ! -e "$SNAP_DATA"/edgex-services-env ]; then
    cp "$SNAP"/config/edgex-services-env "$SNAP_DATA"
fi

. "$SNAP_DATA"/edgex-services-env

if [ "$DEVICE_VIRTUAL" = "y" ] ; then
    pid=$(pgrep -f device-virtual)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "device-virtual"
    fi
fi

if [ "$EXPORT_DISTRO" = "y" ] ; then
    pid=$(pgrep -f  export-distro)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "export-distro"
    fi
fi

if [ "$EXPORT_CLIENT" = "y" ] ; then
    pid=$(pgrep -f export-client)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "export-client"
    fi
fi

if [ "$SUPPORT_SCHEDULER" = "y" ] ; then
    pid=$(pgrep -f support-scheduler)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "support-scheduler"
    fi
fi

if [ "$CORE_COMMAND" = "y" ] ; then
    pid=$(pgrep -f core-command)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "core-command"
    fi
fi

if [ "$CORE_DATA" = "y" ] ; then
    pid=$(pgrep -f core-data)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "core-data"
    fi
fi

if [ "$CORE_METADATA" = "y" ] ; then
    pid=$(pgrep -f core-metadata)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "core-metadata"
    fi
fi

if [ "$SUPPORT_NOTIFICATIONS" = "y" ] ; then
    pid=$(pgrep -f support-notifications)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "support-notifications"
    fi
fi

if [ "$SUPPORT_LOGGING" = "y" ] ; then
    pid=$(pgrep -f support-logging)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	kill_service "$pid" "support-logging"
    fi
fi

if [ "$SECURITY" ]; then
    # stop kong - note this might fail, so we just force it to always succeed with "|| true"
    $SNAP/bin/kong-wrapper.sh stop -p $SNAP_DATA/kong || true

    # killing nginx is tricky, as we need to make sure we don't inadvertantly kill any other nginx
    # processes, so we use egrep with a more specific pattern
    pid=`pgrep -f "nginx.*${SNAP_DATA}/kong"` || true
    if [ ! -z $pid ] && [ $pid != "" ] ; then
        int_service $pid "nginx parent"
    fi

    # finally for kong also remove the kong_env file so that file doesn't stop the start script from working again
    rm -f $SNAP_DATA/kong/.kong_env

    # kill vault-worker, as it might get stuck in an infinite loop trying to unseal the vault
    pid=`pgrep -f "bin/vault-worker.sh"` || true
    if [ ! -z $pid ] && [ $pid != "" ] ; then
        kill_service $pid "vault worker"
    fi

    # send sigint to vault to shut it down
    pid=`pgrep -f "bin/vault.*config/security-secret-store/vault-config.json"` || true
    if [ ! -z $pid ] && [ $pid != "" ] ; then
        int_service $pid "vault"
    fi
fi

if [ "$CORE_DATA" = "y" ] || [ "$CORE_METADATA" = "y" ] ; then
    pid=$(pgrep -f mongod)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	echo "shutting down mongod..."
	"$SNAP/bin/mongod" --shutdown --dbpath "$SNAP_DATA/mongo/db" || true
    fi
fi

pid=$(pgrep -f "consul\ agent")
if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
    echo $pid

    # use SIGINT to gracefully shutdown
    echo "shutting down consul..."
    int_service "$pid" "consul"
fi
