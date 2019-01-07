#!/usr/bin/env sh

# Copyright (c) 2018
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

# Start EdgeX Foundry services in right order, as described:
# https://docs.edgexfoundry.org/Ch-GettingStartedUsers.html

# EDGEX_COMPOSE_FILE overrides the default compose file. 
# EDGEX_CORE_DB identifies the DB for Core Services.
# EDGEX_SERVICES lists the service to start
#
# E.g. Start everything but metadata and data, use Redis for Core Services and a local compose file
# EDGEX_SERVICES="logging command export-client export-distro notifications" \
# EDGEX_CORE_DB=redis EDGEX_COMPOSE_FILE=docker/local-docker-compose.yml \
# bin/edgex-docker-launch.sh

if [ -z $EDGEX_COMPOSE_FILE ]; then
  COMPOSE_FILENAME=docker-compose-delhi-0.7.1.yml
  COMPOSE_FILE=/tmp/${COMPOSE_FILENAME}
  COMPOSE_URL=https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/compose-files/${COMPOSE_FILENAME}
  
  echo "Pulling latest compose file..."
  curl -o $COMPOSE_FILE $COMPOSE_URL
else
  COMPOSE_FILE=$EDGEX_COMPOSE_FILE
fi

EDGEX_CORE_DB=${EDGEX_CORE_DB:-"mongo"}

echo "Starting Mongo"
docker-compose -f $COMPOSE_FILE up -d mongo

if [ ${EDGEX_CORE_DB} != mongo ]; then
  echo "Starting $EDGEX_CORE_DB for Core Data Services"
  docker-compose -f $COMPOSE_FILE up -d $EDGEX_CORE_DB
fi

echo "Starting consul"
docker-compose -f $COMPOSE_FILE up -d consul
echo "Populating configuration"
docker-compose -f $COMPOSE_FILE up -d config-seed

echo "Sleeping before launching remaining services"
sleep 15

defaultServices="logging metadata data command export-client export-distro notifications scheduler"
if [ -z ${EDGEX_SERVICES} ]; then
  deps=
  services=${defaultServices}
else
  deps=--no-deps
  services=${EDGEX_SERVICES}
fi

for s in ${services}; do
    echo Starting ${s}
    docker-compose -f $COMPOSE_FILE up -d ${deps} $s
done
