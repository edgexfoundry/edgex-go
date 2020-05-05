#!/bin/bash -e

"$SNAP"/bin/java -jar -Djava.security.egd=file:/dev/urandom -Xmx100M \
            -Dlogging.file="$SNAP_COMMON"/logs/edgex-support-rulesengine.log \
            "$SNAP"/jar/support-rulesengine/support-rulesengine.jar
