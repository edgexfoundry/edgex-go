#!/bin/bash -e

# get the directory of this script
# snippet from https://stackoverflow.com/a/246128/10102404
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
GIT_ROOT=$(dirname "$SCRIPT_DIR")

EXIT_CODE=0

cd "$GIT_ROOT"

if [ -f go.mod.bk ]; then
    echo "go.mod.bk exits - remove before continuing"
    exit 1
fi

# backup go.mod
cp go.mod go.mod.bk

trap cleanup 1 2 3 6

cleanup()
{
    cd "$GIT_ROOT"
    # restore the go.mod file dir
    rm go.mod
    if [ -f go.mod.bk ]; then
        mv go.mod.bk go.mod
    fi
    exit $EXIT_CODE
}

# if go.mod doesn't exist then fail
if [ ! -f go.mod ]; then
    echo "missing go.mod, please fix"
    EXIT_CODE=1
    cleanup
fi

GO111MODULE=on go mod tidy

# check if go.mod and go.mod.bk are the same

set +e
changes=$(diff -u go.mod go.mod.bk)
set -e

if [ -n "$changes" ]; then
    echo "go.mod is not tidy, please run \"go mod tidy\""
    echo "changes from running \"go mod tidy:\""
    echo "$changes"
    EXIT_CODE=1
fi

cleanup
