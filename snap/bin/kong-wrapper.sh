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

# perl lib paths are needed for some rocks that kong loads through luarocks dependencies
export PERL5LIB="$SNAP/usr/local/lib/$archLibName/perl/5.22.1:$SNAP/usr/local/share/perl/5.22.1:$SNAP/usr/lib/$archLibName/perl5/5.22:$SNAP/usr/share/perl5:$SNAP/usr/lib/$archLibName/perl/5.22:$SNAP/usr/share/perl/5.22:$SNAP/usr/local/lib/site_perl:$SNAP/usr/lib/$archLibName/perl-base"

# lua paths so that luarocks can work
export LUA_PATH="$SNAP/lualib/?.lua;$SNAP/lualib/?/init.lua;$SNAP/usr/share/lua/5.1/?.lua;$SNAP/usr/share/lua/5.1/?/init.lua;$SNAP/lib/lua/5.1/?.lua;$SNAP/lib/lua/5.1/?/init.lua;$SNAP/share/lua/5.1/?.lua;$SNAP/share/lua/5.1/?/init.lua;;"
export LUA_CPATH="$SNAP/lualib/?.so;$SNAP/lib/lua/5.1/?.so;$SNAP/lib/$archLibName/lua/5.1/?.so;;"

exec "$SNAP/bin/kong" "$@"