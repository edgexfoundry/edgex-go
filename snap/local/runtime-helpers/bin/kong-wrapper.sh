#!/bin/sh -e

# need to add luajit/lib folder so that luajit will load properly
export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$SNAP/luajit/lib"

# figure out the snap architecture lib name
case $SNAP_ARCH in
    amd64)
        archLibName="x86_64-linux-gnu"
        ;;
    armhf)
        archLibName="arm-linux-gnueabihf"
        ;;
    arm64)
        archLibName="aarch64-linux-gnu"
        ;;
    i386)
        archLibName="i386-linux-gnu"
        ;;
    *)
        # unsupported or unknown architecture
        exit 1
        ;;
esac

# vars that make perl warnings go away
export LC_ALL=C.UTF-8
export LANG=C.UTF-8

# get the perl version
PERL_VERSION=$(perl -version | grep -Po '\(v\K([^\)]*)')

# perl lib paths are needed for some rocks that kong loads through luarocks dependencies
PERL5LIB="$PERL5LIB:$SNAP/usr/lib/$archLibName/perl/$PERL_VERSION"
PERL5LIB="$PERL5LIB:$SNAP/usr/share/perl/$PERL_VERSION"
export PERL5LIB

# lua paths so that luarocks can work
export LUA_VERSION=5.1
export LUA_PATH="$SNAP/lualib/?.lua;$SNAP/lualib/?/init.lua;$SNAP/usr/share/lua/$LUA_VERSION/?.lua;$SNAP/usr/share/lua/$LUA_VERSION/?/init.lua;$SNAP/lib/lua/$LUA_VERSION/?.lua;$SNAP/lib/lua/$LUA_VERSION/?/init.lua;$SNAP/share/lua/$LUA_VERSION/?.lua;$SNAP/share/lua/$LUA_VERSION/?/init.lua;;"
export LUA_CPATH="$SNAP/lualib/?.so;$SNAP/lib/lua/$LUA_VERSION/?.so;$SNAP/lib/$archLibName/lua/$LUA_VERSION/?.so;;"

"$SNAP/bin/kong" "$@"
