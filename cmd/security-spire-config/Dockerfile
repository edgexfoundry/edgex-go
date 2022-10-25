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
ARG BUILDER_BASE=golang:1.18-alpine3.16
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

RUN apk add --update --no-cache make git build-base curl

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .

ARG SPIRE_RELEASE=1.2.1

# build spire from the source in order to be compatible with arch arm64 as well
# in CI the BUILDER_BASE will already contain a compiled spire-server/agent
# so we check to see if the binary is already in the image before compilation
WORKDIR /edgex-go/spire-build
RUN if ! test -f /usr/local/bin/spire-server; then wget -q "https://github.com/spiffe/spire/archive/refs/tags/v${SPIRE_RELEASE}.tar.gz" && \
    tar xv --strip-components=1 -f "v${SPIRE_RELEASE}.tar.gz" && \
    echo "building spire from source..." && \
    make bin/spire-server bin/spire-agent && \
    cp bin/spire* /usr/local/bin/; \
    fi

WORKDIR /edgex-go

# Deployment image
FROM alpine:3.15

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2022 Intel Corporation'

RUN apk update && apk --no-cache --update add dumb-init gcompat

COPY --from=builder /usr/local/bin/spire-server /usr/local/bin
COPY --from=builder /edgex-go/cmd/security-spire-config/docker-entrypoint.sh /usr/local/bin/

WORKDIR /usr/local/etc/spiffe-scripts.d
COPY --from=builder /edgex-go/cmd/security-spire-config/seed_builtin_entries.sh /usr/local/etc/spiffe-scripts.d

WORKDIR /

ENTRYPOINT [ "/usr/bin/dumb-init" ]
CMD [ "--verbose", "docker-entrypoint.sh" ]
