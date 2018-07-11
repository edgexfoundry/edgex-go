#!/bin/sh

# Copyright (c) 2018
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

# Start EdgeX Foundry services in right order, as described:
# https://wiki.edgexfoundry.org/display/FA/Get+EdgeX+Foundry+-+Users

COMPOSE_FILE=../docker/docker-compose.yml
COMPOSE_URL=https://raw.githubusercontent.com/edgexfoundry/developer-scripts/master/compose-files/docker-compose-california-0.6.0.yml

echo "Pulling latest compose file..."
curl -o $COMPOSE_FILE $COMPOSE_URL

echo "Starting mongo"
docker-compose -f $COMPOSE_FILE up -d mongo
echo "Starting consul"
docker-compose -f $COMPOSE_FILE up -d consul
echo "Populating configuration"
docker-compose -f $COMPOSE_FILE up -d config-seed

echo "Sleeping before launching remaining services"
sleep 15

echo "Starting support-logging"
docker-compose -f $COMPOSE_FILE up -d logging
echo "Starting core-metadata"
docker-compose -f $COMPOSE_FILE up -d metadata
echo "Starting core-data"
docker-compose -f $COMPOSE_FILE up -d data
echo "Starting core-command"
docker-compose -f $COMPOSE_FILE up -d command
echo "Starting core-export-client"
docker-compose -f $COMPOSE_FILE up -d export-client
echo "Starting core-export-distro"
docker-compose -f $COMPOSE_FILE up -d export-distro
