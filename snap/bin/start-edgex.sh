#!/bin/sh
set -ex

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

if [ "$SUPPORT_LOGGING" = "y" ] ; then
    sleep 60
    echo "Starting logging"

    cd "$SNAP"/config/support-logging
    "$SNAP"/bin/support-logging --consul &
fi

if [ "$SUPPORT_NOTIFICATIONS" = "y" ] ; then
    sleep 65
    echo "Starting notifications"

    "$JAVA" -jar -Djava.security.egd=file:/dev/urandom -Xmx100M \
               -Dspring.cloud.consul.enabled=true \
               -Dlogging.file="$SNAP_COMMON"/edgex-notifications.log \
               "$SNAP"/jar/support-notifications/support-notifications.jar &
fi


if [ "$CORE_METADATA" = "y" ] ; then
    sleep 33
    echo "Starting metadata"

    cd "$SNAP"/config/core-metadata
    "$SNAP"/bin/core-metadata --consul &
fi

if [ "$CORE_DATA" = "y" ] ; then
    sleep 60
    echo "Starting core-data"

    cd "$SNAP"/config/core-data
    "$SNAP"/bin/core-data --consul &
fi


if [ "$CORE_COMMAND" = "y" ] ; then
    sleep 60
    echo "Starting command"

    cd "$SNAP"/config/core-command
    "$SNAP"/bin/core-command --consul &
fi


if [ "$SUPPORT_SCHEDULER" = "y" ] ; then
    sleep 60
    echo "Starting scheduler"

    # workaround consul.host=edgex-core-consul bug:
    # https://github.com/edgexfoundry/support-scheduler/issues/23

    "$JAVA" -jar -Djava.security.egd=file:/dev/urandom -Xmx100M \
               -Dspring.cloud.consul.enabled=true \
               -Dspring.cloud.consul.host=localhost \
               -Dlogging.file="$SNAP_COMMON"/edgex-support-scheduler.log \
               "$SNAP"/jar/support-scheduler/support-scheduler.jar &
fi

if [ "$EXPORT_CLIENT" = "y" ] ; then
    sleep 60
    echo "Starting export-client"

    # TODO: fix log file in res/configuration.json
    cd "$SNAP"/config/export-client
    "$SNAP"/bin/export-client --consul &
fi

if [ "$EXPORT_DISTRO" = "y" ] ; then
    sleep 60
    echo "Starting export-distro"

    # TODO: fix log file in res/configuration.json
    cd "$SNAP"/config/export-distro
    "$SNAP"/bin/export-distro --consul &
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
               -Dlogging.file="$SNAP_COMMON"/edgex-device-virtual.log \
               -Dapplication.device-profile-paths="$SNAP_COMMON"/bacnet_profiles,"$SNAP_COMMON"/modbus_profiles \
               "$SNAP"/jar/device-virtual/device-virtual.jar &
fi
