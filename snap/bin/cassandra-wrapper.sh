#!/bin/bash -e

# this script combined from https://github.com/snapcrafters/cassandra/blob/3df209fafb7b3dd341ec7412be4725d7104c5a54/common
# and from https://github.com/snapcrafters/cassandra/blob/3df209fafb7b3dd341ec7412be4725d7104c5a54/wrapper-cassandra

# note: this was changed to force all cassandra data to live inside a cassandra subfolder
# inside SNAP_DATA - this is mainly for separation of data from other edgex services
export CASSANDRA_HOME="$SNAP_DATA/cassandra"
export STATIC_CONF="$SNAP/etc/cassandra"

export CASSANDRA_CONF="$CASSANDRA_HOME/etc"
# also edgex specific change - we put all other logs inside $SNAP_COMMON, so change
# cassandra to put it's logs there too
export CASSANDRA_LOGS="$SNAP_COMMON/logs"
export CASSANDRA_DATA="$SNAP_COMMON/data"

if [ -n "$CLASSPATH" ]; then
    CLASSPATH=$CLASSPATH:$CASSANDRA_CONF
else
    CLASSPATH=$CASSANDRA_CONF
fi

for jar in $SNAP/usr/share/cassandra/lib/*.jar; do
    CLASSPATH="$CLASSPATH:$jar"
done

export CLASSPATH

[ -e "$CASSANDRA_CONF" ] || mkdir -p "$CASSANDRA_CONF"
[ -e "$CASSANDRA_LOGS" ] || mkdir -p "$CASSANDRA_LOGS"
[ -e "$CASSANDRA_DATA" ] || mkdir -p "$CASSANDRA_DATA"

# Required config files. Use the default if one is not provided.
for f in cassandra.yaml logback.xml cassandra-env.sh jvm.options; do
    if [ ! -e "$CASSANDRA_CONF/$f" ]; then
        if [ "$f" = "cassandra-env.sh" ]; then
            # Libraries are in /usr/share/cassandra, not CASSANDRA_HOME
            sed 's,\$CASSANDRA_HOME/lib,$SNAP/usr/share/cassandra/lib,' \
                "$STATIC_CONF/$f" > "$CASSANDRA_CONF/$f"
        else
            cp "$STATIC_CONF/$f" "$CASSANDRA_CONF/$f"
        fi
    fi
done

export cassandra_storagedir="$CASSANDRA_DATA"
export JVM_OPTS="$JVM_OPTS -Dcassandra.config=file://$CASSANDRA_CONF/cassandra.yaml"

# set JAVA_HOME
export JAVA_HOME=$(ls -d $SNAP/usr/lib/jvm/java-1.8.0-openjdk-*)

# The -x bit isn't set on cassandra
/bin/sh $SNAP/usr/sbin/cassandra -R -p "$CASSANDRA_HOME/cassandra.pid"
