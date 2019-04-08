#!/bin/bash -e

# the first argument is the service
# the second argument is the operation, one of "stop", "start", "restart"
# TODO: handle enable/disable when those paramaters are provided

if [ "$#" -ne 2 ]; then
    echo "invalid number of arguments"
fi

case "$#" in
    0)
        echo "the service and the operation must be provided"
        exit 1
        ;;
    1)
        echo "the operation must be provided"
        exit 1
        ;;
    2)
        # correct - do nothing
        ;;
    *)
        echo "too many arguments provided"
        echo "only the service and the operation must be provided"
        exit 1
        ;;
esac


# note this prefix is in edgex-go, as go code, and if it changes we don't have
# a good way to pull that here so if this breaks go look for changes there
svcName=${1#"edgex-"}

# the operation type maps 1:1 to the snapctl service operations
snapctl "$2" "$SNAP_NAME.$svcName"

