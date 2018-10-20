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

$JAVA -jar -Djava.security.egd=file:/dev/urandom -Xmx100M \
            -Dspring.cloud.consul.enabled=true \
            -Dspring.cloud.consul.host=localhost \
            -Dlogging.file=$SNAP_COMMON/logs/edgex-support-rulesengine.log \
            $SNAP/jar/support-rulesengine/support-rulesengine.jar
