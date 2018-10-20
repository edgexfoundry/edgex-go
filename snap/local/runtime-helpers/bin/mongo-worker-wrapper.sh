#!/bin/bash -e


while true; do
  mongo $SNAP/mongo/init_mongo.js && break
  sleep 5
done
