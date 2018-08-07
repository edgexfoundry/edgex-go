#!/bin/sh
set -ex

# TODO: it would be better if this func first tried SIGTERM,
# and fell back to SIGKILL, however this script is not meant
# to be long-lived, hopefully deprecated once everything
# all micro services becomes snap daemons.
kill_service() {
	echo "sending SIGKILL to $2 ($1) service"
	kill -9 $1 || true
}

# send SIGINT to consul
int_service() {
	echo "sending SIGINT to $2 ($1) service"
	kill -INT $1  || true
}

# Bootstrap service env vars
if [ ! -e $SNAP_DATA/edgex-services-env ]; then
    cp $SNAP/config/edgex-services-env $SNAP_DATA
fi

. $SNAP_DATA/edgex-services-env

if [ $DEVICE_VIRTUAL = "y" ] ; then
    pid=`ps -ef | grep device-virtual | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "device-virtual"
    fi
fi

if [ $EXPORT_DISTRO = "y" ] ; then
    pid=`ps -ef | grep export-distro | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "export-distro"
    fi
fi

if [ $EXPORT_CLIENT = "y" ] ; then
    pid=`ps -ef | grep export-client | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "export-client"
    fi
fi

if [ $SUPPORT_SCHEDULER = "y" ] ; then
    pid=`ps -ef | grep support-scheduler | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "support-scheduler"
    fi
fi

if [ $CORE_COMMAND = "y" ] ; then
    pid=`ps -ef | grep core-command | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "core-command"
    fi
fi

if [ $CORE_DATA = "y" ] ; then
    pid=`ps -ef | grep core-data | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "core-data"
    fi
fi

if [ $CORE_METADATA = "y" ] ; then
    pid=`ps -ef | grep core-metadata | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "core-metadata"
    fi
fi

if [ $SUPPORT_NOTIFICATIONS = "y" ] ; then
    pid=`ps -ef | grep support-notifications | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "support-notifications"
    fi
fi

if [ $SUPPORT_LOGGING = "y" ] ; then
    pid=`ps -ef | grep support-logging | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	kill_service $pid "support-logging"
    fi
fi

if [ $CORE_DATA = "y" ] || [ $CORE_METADATA = "y" ] ; then
    pid=`ps -ef | grep mongod | grep -v grep | awk '{print $2}'`

    if [ ! -z $pid ] && [ $pid != "" ] ; then
	echo "shutting down mongod..."
	$SNAP/bin/mongod --shutdown --dbpath $SNAP_DATA/mongo/db || true
    fi
fi

pid=`ps -ef | grep "consul\ agent" | grep -v grep | awk '{print $2}'`
if [ ! -z $pid ] && [ $pid != "" ] ; then

    # use SIGINT to gracefully shutdown
    echo "shutting down consul..."
    int_service $pid "consul"
fi
