#
# Copyright (c) 2018
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

ARG BUILDER_BASE=golang:1.13-alpine
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.
RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make bash git

COPY go.mod .

RUN go mod download

COPY . .

# Build the SMA executable.
RUN make cmd/sys-mgmt-agent/sys-mgmt-agent

# Build the golang "executor" executable.
RUN make cmd/sys-mgmt-executor/sys-mgmt-executor

# Get the Docker-in-Docker image layered-in now.
FROM docker:latest

RUN apk add --no-cache bash
RUN rm -rf /var/cache/apk/*

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2017-2019: Mainflux, Cavium, Dell'

ENV APP_PORT=48090
#expose system mgmt agent port
EXPOSE $APP_PORT

# Copy over the SMA executable bits.
COPY --from=builder /edgex-go/cmd/sys-mgmt-agent/sys-mgmt-agent /
COPY --from=builder /edgex-go/cmd/sys-mgmt-agent/res/configuration.toml /res/configuration.toml

# Copy over the golang "executor" executable.
COPY --from=builder /edgex-go/cmd/sys-mgmt-executor/sys-mgmt-executor /

RUN apk --no-cache add py-pip
RUN pip install docker-compose==1.23.2
RUN apk --no-cache add curl

ENTRYPOINT ["/sys-mgmt-agent","-cp=consul.http://edgex-core-consul:8500","--registry","--confdir=/res"]

# Following (commented-out) line is a life-saving DEBUG statement.
# ENTRYPOINT tail -f /dev/null
