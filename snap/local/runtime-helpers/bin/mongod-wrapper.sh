#!/bin/bash -e

# check the mongo database path
MONGO_DATA_DIR="$SNAP_DATA"/mongo/db
if [ ! -e "$MONGO_DATA_DIR" ] ; then
    mkdir -p "$MONGO_DATA_DIR"
fi

# now start up mongo 
$SNAP/bin/mongod --dbpath $MONGO_DATA_DIR --logpath $SNAP_COMMON/mongodb.log --smallfiles
