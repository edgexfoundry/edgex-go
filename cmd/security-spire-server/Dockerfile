#  ----------------------------------------------------------------------------------
#  Copyright 2022 Intel Corporation
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
# 
#  ----------------------------------------------------------------------------------

# Build utility container
ARG BUILDER_BASE=golang:1.17-alpine3.15
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk add --update --no-cache make git build-base curl

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .

ARG SPIRE_RELEASE=1.2.1

# build spire from the source in order to be compatible with arch arm64 as well
RUN mkdir -p spire-build
WORKDIR /edgex-go/spire-build
RUN wget -q "https://github.com/spiffe/spire/archive/refs/tags/v${SPIRE_RELEASE}.tar.gz" && \
    tar xv --strip-components=1 -f "v${SPIRE_RELEASE}.tar.gz" && \
    echo "building spire from source..." && \
    make bin/spire-server bin/spire-agent

# Deployment image
FROM alpine:3.15

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2022 Intel Corporation'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk update && apk --no-cache --update add dumb-init openssl gcompat

COPY --from=builder /edgex-go/spire-build/bin/spire-server /usr/local/bin
COPY --from=builder /edgex-go/cmd/security-spire-server/docker-entrypoint.sh /usr/local/bin/
COPY --from=builder /edgex-go/cmd/security-spire-server/server.conf /usr/local/etc/spire/server.conf.tpl
COPY --from=builder /edgex-go/cmd/security-spire-server/openssl.conf /usr/local/etc/

ENTRYPOINT [ "/usr/bin/dumb-init" ]
CMD [ "--verbose", "docker-entrypoint.sh" ]
