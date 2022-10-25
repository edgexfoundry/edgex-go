#
# Copyright (c) 2018
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

ARG BUILDER_BASE=golang:1.18-alpine3.16
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

RUN apk add --update --no-cache make bash git

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
# Build the SMA executable.
RUN make cmd/sys-mgmt-agent/sys-mgmt-agent

# Build the golang "executor" executable.
RUN make cmd/sys-mgmt-executor/sys-mgmt-executor

# Get the Docker-in-Docker image layered-in now.
FROM alpine:3.16

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2017-2019: Mainflux, Cavium, Dell, Copyright (c) 2021: Intel Corporation'

# consul token needs to be security-bootstrappable and for security-bootstrappable, dumb-init is required
RUN apk add --update --no-cache bash dumb-init py3-pip curl docker-cli && \
      pip install --no-cache-dir docker-compose==1.23.2

ENV APP_PORT=58890
#expose system mgmt agent port
EXPOSE $APP_PORT

# Copy over the SMA executable bits.
COPY --from=builder /edgex-go/cmd/sys-mgmt-agent/sys-mgmt-agent /
COPY --from=builder /edgex-go/cmd/sys-mgmt-agent/res/configuration.toml /res/configuration.toml

# Copy over the golang "executor" executable.
COPY --from=builder /edgex-go/cmd/sys-mgmt-executor/sys-mgmt-executor /

ENTRYPOINT ["/sys-mgmt-agent"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]

# Following (commented-out) line is a life-saving DEBUG statement.
# ENTRYPOINT tail -f /dev/null
