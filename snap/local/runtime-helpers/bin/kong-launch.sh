#!/bin/sh -e

# need to add luajit/lib folder so that luajit will load properly
export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$SNAP/luajit/lib"

# lua paths so that luarocks can work
export LUA_VERSION=5.1
export LUA_PATH="$SNAP/lualib/?.lua;$SNAP/lualib/?/init.lua;$SNAP/usr/share/lua/$LUA_VERSION/?.lua;$SNAP/usr/share/lua/$LUA_VERSION/?/init.lua;$SNAP/lib/lua/$LUA_VERSION/?.lua;$SNAP/lib/lua/$LUA_VERSION/?/init.lua;$SNAP/share/lua/$LUA_VERSION/?.lua;$SNAP/share/lua/$LUA_VERSION/?/init.lua;;"
export LUA_CPATH="$SNAP/lualib/?.so;$SNAP/lib/lua/$LUA_VERSION/?.so;$SNAP/lib/$ARCH_LIB_NAME/lua/$LUA_VERSION/?.so;;"

# set postgresql password
export KONG_PG_PASSWORD=`cat "$SNAP_DATA/config/postgres/kongpw"`

# also make sure our logs dir exists because for some reason nginx can't create
# it's own logs dir
mkdir -p "$KONG_LOGS_DIR"

exec "$@"
