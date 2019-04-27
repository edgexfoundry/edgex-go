#
# Copyright (c) 2018
# Dell
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

# Docker image for Golang Core Data micro service 
FROM golang:1.11-alpine AS builder

ENV GO111MODULE=on
WORKDIR /go/src/github.com/edgexfoundry/edgex-go

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add zeromq-dev libsodium-dev pkgconfig build-base git

COPY go.mod .
#COPY go.sum .

RUN go mod download

COPY . .

RUN make core-data

#Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Dell, Cavium'

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk --no-cache add zeromq
COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/core-data/Attribution.txt /
COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/core-data/core-data /
COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/core-data/res/docker/configuration.toml /res/docker/configuration.toml

ENTRYPOINT ["/core-data","--registry","--profile=docker","--confdir=/res"]
