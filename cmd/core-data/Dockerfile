#
# Copyright (c) 2018
# Dell
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

# Docker image for Golang Core Data micro service
ARG BUILDER_BASE=golang:1.13-alpine
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add zeromq-dev libsodium-dev pkgconfig build-base git

COPY go.mod .

RUN go mod download

COPY . .
RUN make cmd/core-data/core-data

#Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Dell, Cavium'

ENV APP_PORT=48080
#expose core data port
EXPOSE $APP_PORT

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.
RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk --no-cache add zeromq
COPY --from=builder /edgex-go/cmd/core-data/Attribution.txt /
COPY --from=builder /edgex-go/cmd/core-data/core-data /
COPY --from=builder /edgex-go/cmd/core-data/res/configuration.toml /res/configuration.toml

ENTRYPOINT ["/core-data"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]
