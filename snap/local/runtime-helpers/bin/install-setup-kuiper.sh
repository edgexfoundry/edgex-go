#!/bin/sh

if [ ! -f "$SNAP_DATA/kuiper/etc" ]; then
    mkdir -p "$SNAP_DATA/kuiper/etc"
fi

cp -rv "$SNAP_DATA/etc" "$SNAP_DATA/kuiper"

