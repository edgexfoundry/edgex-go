#!/bin/bash
#
# Copyright (c) 2019
# Mainflux
#
# SPDX-License-Identifier: Apache-2.0


# Collects EdgeX Go binaries (must be previously built)
# and packages them into a dir alongside with necessary config files.
# This facililtets transfer of binaries to otehr machine
# (for example copy of cross-compiled binaris to ARM gateway)

DIR=$PWD
CMD_DIR=../cmd
SERVICES=(config-seed export-client core-metadata core-command support-logging \
  support-notifications sys-mgmt-executor sys-mgmt-agent support-scheduler \
  core-data export-distro)

PKG_DIR=${1:-edgex}

# Copy binaries and config
mkdir -p $PKG_DIR
for i in "${SERVICES[@]}"; do
  mkdir -p $PKG_DIR/$i
  cp $CMD_DIR/$i/$i $PKG_DIR/$i
  cp $CMD_DIR/$i/res/configuration.toml $PKG_DIR/$i 2>/dev/null || :
done

# Copy and modify run script
cp run.sh $PKG_DIR
sed -i 's#BIN_DIR=../cmd#BIN_DIR=.#g' $PKG_DIR/run.sh
sed -i 's#EDGEX_CONF_DIR=./res#EDGEX_CONF_DIR=.#g' $PKG_DIR/run.sh