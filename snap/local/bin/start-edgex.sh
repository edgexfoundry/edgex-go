#!/bin/sh
set -ex

cd $SNAP_DATA

if [ "$(arch)" = "aarch64" ] ; then
    ARCH="arm64"
elif [ "$(arch)" = "x86_64" ] ; then
    ARCH="amd64"
else
    echo "Unsupported architecture: $(arch)"
    exit 1
fi

JAVA="$SNAP"/usr/lib/jvm/java-8-openjdk-"$ARCH"/jre/bin/java

# Bootstrap service env vars
if [ ! -e "$SNAP_DATA"/edgex-services-env ]; then
    cp "$SNAP"/config/edgex-services-env "$SNAP_DATA"
fi

. "$SNAP_DATA"/edgex-services-env

echo "Starting config-registry (consul)..."
"$SNAP"/bin/start-consul.sh

sleep 60

MONGO_DATA_DIR="$SNAP_DATA"/mongo/db

echo "Starting config-seed..."
"$SNAP/bin/start-config-seed.sh"

echo "Starting mongo..."
if [ -e "$MONGO_DATA_DIR" ] ; then
    rm -rf "${MONGO_DATA_DIR:?}"/*
else
    mkdir -p "$MONGO_DATA_DIR"
fi

"$SNAP"/mongo/launch-edgex-mongo.sh

if [ "$SECURITY" = "y" ] ; then
    echo "Starting up security services"
    $SNAP/bin/security-start.sh
fi

if [ "$SUPPORT_LOGGING" = "y" ] ; then
    sleep 60
    echo "Starting logging"

    cd $SNAP_DATA/config/support-logging
    $SNAP/bin/support-logging --consul &
fi

if [ "$SUPPORT_NOTIFICATIONS" = "y" ] ; then
    sleep 65
    echo "Starting notifications"

    cd $SNAP_DATA/config/support-notifications
    $SNAP/bin/support-notifications --consul &
fi


if [ "$CORE_METADATA" = "y" ] ; then
    sleep 33
    echo "Starting metadata"

    cd $SNAP_DATA/config/core-metadata
    $SNAP/bin/core-metadata --consul &
fi

if [ "$CORE_DATA" = "y" ] ; then
    sleep 60
    echo "Starting core-data"

    cd $SNAP_DATA/config/core-data
    $SNAP/bin/core-data --consul &
fi


if [ "$CORE_COMMAND" = "y" ] ; then
    sleep 60
    echo "Starting command"

    cd $SNAP_DATA/config/core-command
    $SNAP/bin/core-command --consul &
fi


if [ "$SUPPORT_SCHEDULER" = "y" ] ; then
    sleep 60
    echo "Starting scheduler"
    cd $SNAP_DATA/config/support-scheduler
    $SNAP/bin/support-scheduler --consul &
fi

if [ "$EXPORT_CLIENT" = "y" ] ; then
    sleep 60
    echo "Starting export-client"

    # TODO: fix log file in res/configuration.json
    cd $SNAP_DATA/config/export-client
    $SNAP/bin/export-client --consul &
fi

if [ "$EXPORT_DISTRO" = "y" ] ; then
    sleep 60
    echo "Starting export-distro"

    # TODO: fix log file in res/configuration.json
    cd $SNAP_DATA/config/export-distro
    $SNAP/bin/export-distro --consul &
fi

if [ "$DEVICE_VIRTUAL" = "y" ] ; then
    sleep 60
    echo "Starting device-virtual"

    # first-time, create sample profile dirs in $SNAP_COMMON
    if [ ! -e "$SNAP_COMMON"/bacnet_profiles ]; then
	mkdir "$SNAP_COMMON"/bacnet_profiles
	cp "$SNAP"/jar/device-virtual/bacnet_sample_profiles/*.yaml \
	   "$SNAP_COMMON"/bacnet_profiles
    fi

    if [ ! -e "$SNAP_COMMON"/modbus_profiles ]; then
	mkdir "$SNAP_COMMON"/modbus_profiles
	cp "$SNAP"/jar/device-virtual/modbus_sample_profiles/*.yaml \
	   "$SNAP_COMMON"/modbus_profiles
    fi

    cd "$SNAP"/jar/device-virtual
    "$JAVA" -jar -Djava.security.egd=file:/dev/urandom -Xmx100M \
               -Dspring.cloud.consul.enabled=false \
               -Dlogging.level.org.edgexfoundry=DEBUG \
               -Dlogging.file=$SNAP_COMMON/logs/edgex-device-virtual.log \
               -Dapplication.device-profile-paths=$SNAP_COMMON/bacnet_profiles,$SNAP_COMMON/modbus_profiles \
               $SNAP/jar/device-virtual/device-virtual.jar &
fi
