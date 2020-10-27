#!/bin/sh -e

OPENRESTY_PATH="/usr/local/openresty/"

# need to add luajit/lib folder so that luajit will load properly
export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$OPENRESTY_PATH/luajit/lib"

# lua paths so that luarocks can work
export LUA_VERSION=5.1
# lua paths so that luarocks can work
export LUA_VERSION=5.1
export LUA_PATH="$OPENRESTY_PATH/lualib/?.lua;$OPENRESTY_PATH/lualib/?/init.lua;/usr/share/lua/$LUA_VERSION/?.lua;/usr/share/lua/$LUA_VERSION/?/init.lua;/usr/local/lib/lua/$LUA_VERSION/?.lua;/usr/local/lib/lua/$LUA_VERSION/?/init.lua;/usr/local/share/lua/$LUA_VERSION/?.lua;/usr/local/share/lua/$LUA_VERSION/?/init.lua;;"
export LUA_CPATH="$OPENRESTY_PATH/lualib/?.so;/usr/local/lib/lua/$LUA_VERSION/?.so;;"

# set postgresql password
export KONG_PG_PASSWORD=`cat "$SNAP_DATA/config/postgres/kongpw"`

# also make sure our logs dir exists because for some reason nginx can't create
# it's own logs dir
mkdir -p "$KONG_LOGS_DIR"

exec "$@"
