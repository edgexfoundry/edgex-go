#
# Copyright (c) 2017-2018
# Mainflux
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.11.2-alpine3.7 AS builder

ENV GO111MODULE=on
WORKDIR /go/src/github.com/edgexfoundry/edgex-go

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.
RUN echo http://nl.alpinelinux.org/alpine/v3.7/main > /etc/apk/repositories; \
    echo http://nl.alpinelinux.org/alpine/v3.7/community >> /etc/apk/repositories


RUN apk update && apk add make git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN make cmd/export-client/export-client

FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2017-2018: Mainflux, Cavium'

COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/export-client/export-client /
COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/export-client/res/docker/configuration.toml /res/docker/configuration.toml
ENTRYPOINT ["/export-client","--consul","--profile=docker","--confdir=/res"]
