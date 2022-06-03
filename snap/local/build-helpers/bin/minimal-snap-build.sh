#!/bin/sh
#
# This script is used by the LF's CI/CD build pipeline to
# optimize the snap CI check run for pull requests. When
# run, it essentially strips out everything (apps and
# parts) from the snapcraft.yaml file except those required
# to build edgex-go, as the whole idea of the CI check is to
# catch changes to edgex-go that break the snap build.
#
# Note - in addition to applying this patch, the pipeline also
# only primes the snap (e.g. `snapcraft prime`), as there's no
# need to build the finally binary .snap file (and it won't
# work with the patch applied). This further reduces the build
# time.

# get the 4.6.2 release of yq from github
curl -L --output yq.tar.gz https://github.com/mikefarah/yq/releases/download/v4.6.2/yq_linux_amd64.tar.gz
tar -xvf yq.tar.gz
mv yq_linux_amd64 yq

CURRDIR=$(pwd)
SNAPCRAFT_YAML="$CURRDIR/snap/snapcraft.yaml"

# remove first chunk of apps
$CURRDIR/yq e -P -i 'del(.apps.consul,.apps.redis,.apps.postgres,.apps.kong-daemon,.apps.vault,.apps.vault-cli)' "$SNAPCRAFT_YAML"

# remove second chunk of apps
$CURRDIR/yq e -P -i 'del(.apps.app-service-configurable)' "$SNAPCRAFT_YAML"

# remove third chunk of apps
$CURRDIR/yq e -P -i 'del(.apps.redis-cli,.apps.consul-cli)' "$SNAPCRAFT_YAML"

# remove fourth chunk of apps
$CURRDIR/yq e -P -i 'del(.apps.kong,.apps.psql,.apps.psql-any,.apps.createdb,.apps.kuiper,.apps.kuiper-cli)' "$SNAPCRAFT_YAML"

# remove unwanted parts
$CURRDIR/yq e -P -i 'del(.parts.snapcraft-preload,.parts.postgres,.parts.consul,.parts.redis,.parts.kong,.parts.vault,.parts.kuiper,.parts.app-service-config)' "$SNAPCRAFT_YAML"


