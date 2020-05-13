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
            -Dlogging.file="$SNAP_COMMON"/logs/edgex-support-rulesengine.log \
            -Drules.default.path="$SNAP_DATA"/support-rulesengine/rules \
            -Drules.template.path="$SNAP_DATA"/support-rulesengine/templates \
            "$SNAP/jar/support-rulesengine/support-rulesengine.jar"
