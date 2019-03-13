#!/bin/bash

# try to initialize mongo, giving up after a reasonable number of tries
MAX_TRIES=10
num_tries=0
until mongo $SNAP/mongo/init_mongo.js; do
    sleep 5
    # increment number of tries
    num_tries=$((num_tries+1))
    if (( num_tries >= MAX_TRIES )); then
        echo "unable to init mongod from mongo-worker after $MAX_TRIES attempts"
        exit 1
    fi
done
