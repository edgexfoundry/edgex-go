#!/bin/sh

# This script is called by the install hook (install.go) during a new install. 
# It configures Postgres for Kong

$SNAP/bin/kong-postgres-setup.sh "install"
