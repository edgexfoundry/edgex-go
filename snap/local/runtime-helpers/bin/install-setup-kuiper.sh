#!/bin/sh

# create kuiper directories in $SNAP_DATA
if [ ! -f "$SNAP_DATA/kuiper/data" ]; then
    mkdir -p "$SNAP_DATA/kuiper/data"
    mkdir -p "$SNAP_DATA/kuiper/etc/functions"
    mkdir -p "$SNAP_DATA/kuiper/etc/multilingual"
    mkdir -p "$SNAP_DATA/kuiper/etc/services"
    mkdir -p "$SNAP_DATA/kuiper/etc/sinks"
    mkdir -p "$SNAP_DATA/kuiper/etc/sources"
    mkdir -p "$SNAP_DATA/kuiper/etc/connections"
    mkdir -p "$SNAP_DATA/kuiper/plugins/functions"
    mkdir -p "$SNAP_DATA/kuiper/plugins/sinks"
    mkdir -p "$SNAP_DATA/kuiper/plugins/sources"
    mkdir -p "$SNAP_DATA/kuiper/plugins/portable"

    for cfg in client kuiper; do
        cp "$SNAP/etc/$cfg.yaml" "$SNAP_DATA/kuiper/etc"
    done

    # Only include the plugin metadata file for mqtt_source,
    # as EdgeX currently doesn't provide a default MQTT broker.
    # Even if it did, configuration (including security) would
    # need to be provided by configuration file (!compliant).
    cp "$SNAP/etc/mqtt_source.json" "$SNAP_DATA/kuiper/etc"

    cp "$SNAP/etc/functions/"*.json "$SNAP_DATA/kuiper/etc/functions"

    cp "$SNAP/etc/services/"*.proto "$SNAP_DATA/kuiper/etc/services"

    for sink in file edgex influx log nop mqtt; do
        cp "$SNAP/etc/sinks/$sink.json" "$SNAP_DATA/kuiper/etc/sinks"
    done

    for src in edgex; do
        cp "$SNAP/etc/sources/$src.json" "$SNAP_DATA/kuiper/etc/sources"
        cp "$SNAP/etc/sources/$src.yaml" "$SNAP_DATA/kuiper/etc/sources"
    done

    cp "$SNAP/etc/connections/connection.yaml" "$SNAP_DATA/kuiper/etc/connections"

    cp "$SNAP/etc/multilingual/"*.ini "$SNAP_DATA/kuiper/etc/multilingual"
fi
