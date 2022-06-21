#!/bin/bash -e

# get the directory of this script
# snippet from https://stackoverflow.com/a/246128/10102404
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
GIT_ROOT=$(dirname "$SCRIPT_DIR")

EXIT_CODE=0

cd "$GIT_ROOT"

if [ ! -f Attribution.txt ]; then
    echo "An Attribution.txt file is missing, please add"
    EXIT_CODE=1
else
    # loop over every library in the go.mod file
    while IFS= read -r lib; do
        if ! grep -q "$lib" Attribution.txt; then
            echo "An attribution for $lib is missing from Attribution.txt, please add"
            # need to do this in a bash subshell, see SC2031
            (( EXIT_CODE=1 ))
        fi
    done < <(cat $GIT_ROOT/go.mod | grep -v 'module ' | grep -v TODO | grep '/' | sed 's/require //' | sed 's/replace //' | awk '{print $1}')
fi
