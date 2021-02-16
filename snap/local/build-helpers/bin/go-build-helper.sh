#!/bin/bash
#
# $1 - go import path
#
# example usage:
# $ gopartbootstrap github.com/edgexfoundry/edgex-go
gopartbootstrap() 
{
    # first set the GOPATH to be in the current directory and in ".gopath"
    GOPATH="$(pwd)/.gopath"
    export GOPATH

    # setup path to include both $SNAPCRAFT_STAGE/bin and $GOPATH/bin
    # the former is for the go tools, as well as things like glide, etc.
    # while the later is for govendor, etc. and other go tools that might need to be installed
    export PATH="$GOPATH/bin:$PATH"

    # now setup the GOPATH for this part using the import path
    export GOIMPORTPATH="$GOPATH/src/$1"
    mkdir -p "$GOIMPORTPATH"
    # note that some tools such as govendor don't work well with symbolic links, so while it's unfortunate
    # we have to copy all this it's a necessary evil at the moment...
    # but note that we do ignore all files that start with "." with the "./*" pattern
    cp -r ./* "$GOIMPORTPATH"

    # finally go into the go import path to prepare for building
    cd "$GOIMPORTPATH" || exit
}
