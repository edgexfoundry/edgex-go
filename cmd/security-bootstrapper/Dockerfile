#  ----------------------------------------------------------------------------------
#  Copyright 2021 Intel Corp.
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
#
#  SPDX-License-Identifier: Apache-2.0
#  ----------------------------------------------------------------------------------

ARG BUILDER_BASE=golang:1.15-alpine3.12
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk add --update --no-cache make git

COPY go.mod .

RUN go mod download

COPY . .

ARG BUILD_BASE_DIR="/tmp/build/dist/linux"

RUN make cmd/security-bootstrapper/security-bootstrapper && \
    make cmd/security-bootstrap-redis/security-bootstrap-redis && \
    # build "dockerize" executable based on the architecture that is currently running on
    set -eux; \
    mkdir -p "$BUILD_BASE_DIR" && \
    cd "$BUILD_BASE_DIR" && \
    git clone https://github.com/jwilder/dockerize.git . && \
    go get github.com/robfig/glock && \
    glock sync -n < GLOCKFILE && \
    TAG=$(git describe --abbrev=0 --tags) && \
    LDFLAGS="-X main.buildVersion=$TAG" && \
    arch="$(apk --print-arch)"; \
    case "$arch" in \
        x86_64   ) \
            echo "building [dockerize] for amd64 ... "; \
            GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -a -tags netgo -installsuffix netgo \
                    -o "$BUILD_BASE_DIR"/dockerize ;; \
        aarch64  ) \
            echo "building [dockerize] for arm64 ... "; \
            GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$BUILD_BASE_DIR"/dockerize ;; \
        *) echo >&2 "Error: dockerize is not supported in arch $arch"; exit 1 ;;\
    esac && \
    echo "$(date) dockerize executable build done and output at directory: $BUILD_BASE_DIR"

FROM alpine:3.12

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2021 Intel Corporation'

RUN apk add --update --no-cache dumb-init openssl su-exec

ENV SECURITY_INIT_DIR /edgex-init
ARG BOOTSTRAP_REDIS_DIR=${SECURITY_INIT_DIR}/bootstrap-redis

RUN mkdir -p ${SECURITY_INIT_DIR} \
    && mkdir -p ${BOOTSTRAP_REDIS_DIR}

WORKDIR ${SECURITY_INIT_DIR}

# Use dockerize utility for services to wait for certain ports to be available
COPY --from=builder /tmp/build/dist/linux/dockerize .

# copy all entrypoint scripts into shared folder
COPY --from=builder /edgex-go/cmd/security-bootstrapper/entrypoint-scripts/ ${SECURITY_INIT_DIR}/
RUN chmod +x ${SECURITY_INIT_DIR}/*.sh

COPY --from=builder /edgex-go/Attribution.txt /
COPY --from=builder /edgex-go/cmd/security-bootstrapper/security-bootstrapper .
COPY --from=builder /edgex-go/cmd/security-bootstrapper/res/configuration.toml ./res/

# needed for bootstrapping Redis db
COPY --from=builder /edgex-go/cmd/security-bootstrap-redis/security-bootstrap-redis ${BOOTSTRAP_REDIS_DIR}/
COPY --from=builder /edgex-go/cmd/security-bootstrap-redis/res/configuration.toml ${BOOTSTRAP_REDIS_DIR}/res/

# Expose the file directory as a volume since there's long-running state
VOLUME ${SECURITY_INIT_DIR}

# setup entry point script
COPY --from=builder /edgex-go/cmd/security-bootstrapper/entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

# gate is one subcommand for security-bootstrapper to do security bootstrapping
CMD ["gate"]
