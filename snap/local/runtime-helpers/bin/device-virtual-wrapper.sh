#!/bin/bash -e

if [ "$(arch)" = "aarch64" ] ; then
    ARCH="arm64"
elif [ "$(arch)" = "x86_64" ] ; then
    ARCH="amd64"
else
    echo "Unsupported architecture: $(arch)"
    exit 1
fi

JAVA="$SNAP"/usr/lib/jvm/java-8-openjdk-"$ARCH"/jre/bin/java

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

# wait for consul to come up
$SNAP/bin/wait-for-consul.sh "device-virtual"

cd "$SNAP"/jar/device-virtual
"$JAVA" -jar -Djava.security.egd=file:/dev/urandom -Xmx100M \
            -Dspring.cloud.consul.enabled=false \
            -Dlogging.level.org.edgexfoundry=DEBUG \
            -Dlogging.file=$SNAP_COMMON/logs/edgex-device-virtual.log \
            -Dapplication.device-profile-paths=$SNAP_COMMON/bacnet_profiles,$SNAP_COMMON/modbus_profiles \
            $SNAP/jar/device-virtual/device-virtual.jar
