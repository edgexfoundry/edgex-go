#!/bin/sh
set -ex

# TODO: it would be better if this func first tried SIGTERM,
# and fell back to SIGKILL, however this script is not meant
# to be long-lived, hopefully deprecated once everything
# all micro services becomes snap daemons.
kill_service() {
	echo "sending SIGKILL to $2 ($1) service"
	kill -9 "$1"
}

# send SIGINT to consul
int_service() {
	echo "sending SIGINT to $2 ($1) service"
	kill -INT "$1"
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

if [ "$CORE_DATA" = "y" ] || [ "$CORE_METADATA" = "y" ] ; then
    pid=$(pgrep -f mongod)

    if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
	echo "shutting down mongod..."
	"$SNAP"/bin/mongod --shutdown --dbpath "$SNAP_DATA"/mongo/db
    fi
fi

pid=$(pgrep -f "consul\ agent")
if [ ! -z "$pid" ] && [ "$pid" != "" ] ; then
    echo $pid

    # use SIGINT to gracefully shutdown
    echo "shutting down consul..."
    int_service "$pid" "consul"
fi
