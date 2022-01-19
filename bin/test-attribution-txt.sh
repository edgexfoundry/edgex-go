#!/bin/bash -e

# get the directory of this script
# snippet from https://stackoverflow.com/a/246128/10102404
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
GIT_ROOT=$(dirname "$SCRIPT_DIR")

EXIT_CODE=0

cd "$GIT_ROOT"

if [ -d vendor.bk ]; then
    echo "vendor.bk exits - remove before continuing"
    exit 1
fi

trap cleanup 0 1 2 3 6

cleanup()
{
    cd "$GIT_ROOT"
    # restore the vendor dir
    rm -rf vendor
    if [ -d vendor.bk ]; then
        mv vendor.bk vendor
    fi
    exit $EXIT_CODE
}

# if the vendor directory exists, back it up so we can build a fresh one
if [ -d vendor ]; then
    mv vendor vendor.bk
fi

# create a vendor dir with the mod dependencies
GO111MODULE=on go mod vendor

# turn on nullglobbing so if there is nothing in cmd dir then we don't do
# anything in this loop
shopt -s nullglob

if [ ! -f Attribution.txt ]; then
    echo "An Attribution.txt file is missing, please add"
    EXIT_CODE=1
else
    # loop over every library in the modules.txt file in vendor
    while IFS= read -r lib; do
        if ! grep -q "$lib" Attribution.txt && [ "$lib" != "explicit" ] && [ "$lib" != "explicit;" ]; then
            echo "An attribution for $lib is missing from Attribution.txt, please add"
            # need to do this in a bash subshell, see SC2031
            (( EXIT_CODE=1 ))
        fi
    done < <(grep '#' < "$GIT_ROOT/vendor/modules.txt" | awk '{print $2}')
fi
