#!/bin/sh -e

# figure out the snap architecture lib name
case $SNAP_ARCH in
    amd64)
        ARCH_LIB_NAME="x86_64-linux-gnu"
        ;;
    arm64)
        ARCH_LIB_NAME="aarch64-linux-gnu"
        ;;
    *)
        # unsupported or unknown architecture
        exit 1
        ;;
esac

export ARCH_LIB_NAME

# get the perl version
PERL_VERSION=$(perl -version | grep -Po '\(v\K([^\)]*)')

# perl lib paths are needed for some rocks that kong loads through luarocks dependencies
PERL5LIB="$PERL5LIB:$SNAP/usr/lib/$ARCH_LIB_NAME/perl/$PERL_VERSION"
PERL5LIB="$PERL5LIB:$SNAP/usr/share/perl/$PERL_VERSION"
PERL5LIB="$PERL5LIB:$SNAP/usr/share/perl5"

export PERL5LIB

exec "$@"
