#!/bin/bash -e

# get the directory of this script
# snippet from https://stackoverflow.com/a/246128/10102404
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
GIT_ROOT=$(dirname "$SCRIPT_DIR")

EXIT_CODE=0

# turn on nullglobbing so if there is nothing in cmd dir then we don't do
# anything in the loops over cmd
shopt -s nullglob

cd "$GIT_ROOT"

if [ -d vendor.bk ]; then
    echo "vendor.bk exits - remove before continuing"
    exit 1
fi

trap cleanup 1 2 3 6

cleanup()
{
    cd "$GIT_ROOT"
    # restore the vendor dir
    rm -r vendor
    if [ -d vendor.bk ]; then
        mv vendor.bk vendor
    fi
    # remove any attribution dependency files from each of the cmd dirs
    for cmd in cmd/* ; do
        cd "$GIT_ROOT/$cmd"
        rm -f dependency-*.bk
    done
    exit $EXIT_CODE
}

# if the vendor directory exists, back it up so we can build a fresh one
if [ -d vendor ]; then
    mv vendor vendor.bk
fi

# create a vendor dir with the mod dependencies
GO111MODULE=on go mod vendor

for cmd in cmd/* ; do
    cd "$cmd"
    if [ ! -f Attribution.txt ]; then
        echo "An Attribution.txt file for $cmd is missing, please add"
        EXIT_CODE=1
    else
        # loop over every library in the modules.txt file in vendor
        while IFS= read -r lib; do
            if ! grep -q "$lib" Attribution.txt; then 
                echo "An attribution for $lib is missing from in $cmd Attribution.txt, please add"
                # need to do this in a bash subshell, see SC2031
                (( EXIT_CODE=1 ))
            fi
        done < <(grep '#' < "$GIT_ROOT/vendor/modules.txt" | awk '{print $2}')

        # now check for stale, unused attributions
        # we need to split the attribution.txt file by blank lines and put each
        # dependency segment into it's own file for processing
        tail -n +2 < Attribution.txt | awk -v RS= '{print > ("dependency-" NR ".bk")}' 
        for dep in ./dependency-*.bk; do
            # the format of the dependency is supposed to be
            # PKG/NAME (LICENSE-TYPE) LIBRARY-LINK
            # LICENSE-LINK
            depName=$(head -n 1 "$dep" | awk '{print $1}')
            if [ -z "$depName" ]; then
                echo "invalid dependency format for $cmd/Attribution.txt"
                EXIT_CODE=1
                break
            else
                # check if the name of the library is in the modules, note that
                # we don't use the link since the link could be ambiguous, but
                # library names should always be unique
                if ! grep "#" < "$GIT_ROOT/vendor/modules.txt" | awk '{print $2}' | grep -q "$depName"; then
                    echo "Unnecessary attribution of $depName in $cmd/Attribution.txt, please remove"
                    EXIT_CODE=1
                fi
            fi
        done
    fi
    cd "$GIT_ROOT"
done

cleanup 
