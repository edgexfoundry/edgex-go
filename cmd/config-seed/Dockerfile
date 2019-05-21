###############################################################################
# Copyright 2018 Dell Technologies, Inc All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###############################################################################

# Docker image for building EdgeX Foundry Config Seed
FROM golang:1.11-alpine AS build-env

# environment variables
ENV GO111MODULE=on
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

# set the working directory
WORKDIR $GOPATH/src/github.com/edgexfoundry/edgex-go

RUN apk update && apk add make git python3 jq

# install remarshal for processing config files in proper way, i.e. turn toml
# into json, use jq to set the correct keys, then turn back into toml in the 
# files
RUN python3 -m pip install remarshal

# copy go source files
COPY go.mod .
#COPY go.sum .

RUN go mod download

COPY . .

# create an identical dir structure under /cmd-redis for using redis
# instead of mongodb
COPY ./cmd ./cmd-redis

# for svcs supporting redis directly, change the following keys before copying
# Databases.Primary.Type = redisdb 
# Databases.Primary.Port = 6379
# Databases.Primary.Host = edgex-redis
RUN for svc in core-data core-metadata export-client support-notifications support-scheduler; do \
    configFile=./cmd-redis/$svc/res/docker/configuration.toml && \
    toml2json --preserve-key-order "$configFile" | \
    jq -r '.Databases.Primary.Type = "redisdb" | .Databases.Primary.Port = 6379 | .Databases.Primary.Host = "edgex-redis"' | \
    json2toml --preserve-key-order > "$configFile.tmp" && \
    mv "$configFile.tmp" "$configFile"; done

# support-logging needs the Writable.Persistence key set to file to be usable with Redis
RUN configFile=./cmd-redis/support-logging/res/docker/configuration.toml && \
    toml2json --preserve-key-order "$configFile" | \
    jq -r '.Writable.Persistence = "file"' | \
    json2toml --preserve-key-order > "$configFile.tmp" && \
    mv "$configFile.tmp" "$configFile"

RUN python3 -m pip uninstall -y remarshal && apk del python3 jq

# build
RUN make cmd/config-seed/config-seed

# Consul Docker image for EdgeX Foundry
FROM golang:1.11-alpine

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2017: Samsung'

RUN apk add --no-cache bash

# environment variables
ENV APP_DIR=/edgex/cmd/config-seed

# set the working directory
WORKDIR $APP_DIR

# copy files
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/config-seed/Attribution.txt  .
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/config-seed/config-seed  .
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/config-seed/res ./res

# copy all the default (i.e. mongodb) configuration.toml files into /edgex/cmd
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/core-command/res /edgex/cmd/core-command/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/core-data/res /edgex/cmd/core-data/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/core-metadata/res /edgex/cmd/core-metadata/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/export-client/res /edgex/cmd/export-client/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/export-distro/res /edgex/cmd/export-distro/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/support-logging/res /edgex/cmd/support-logging/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/support-notifications/res /edgex/cmd/support-notifications/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/support-scheduler/res /edgex/cmd/support-scheduler/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd/sys-mgmt-agent/res /edgex/cmd/sys-mgmt-agent/res

# copy all the redis configuration.toml into /edgex/cmd-redis
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/core-command/res /edgex/cmd-redis/core-command/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/core-data/res /edgex/cmd-redis/core-data/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/core-metadata/res /edgex/cmd-redis/core-metadata/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/export-client/res /edgex/cmd-redis/export-client/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/export-distro/res /edgex/cmd-redis/export-distro/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/support-logging/res /edgex/cmd-redis/support-logging/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/support-notifications/res /edgex/cmd-redis/support-notifications/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/support-scheduler/res /edgex/cmd-redis/support-scheduler/res
COPY --from=build-env $GOPATH/src/github.com/edgexfoundry/edgex-go/cmd-redis/sys-mgmt-agent/res /edgex/cmd-redis/sys-mgmt-agent/res

ENTRYPOINT ["/edgex/cmd/config-seed/config-seed"]
CMD ["--profile=docker", "--cmd=/edgex/cmd"]
