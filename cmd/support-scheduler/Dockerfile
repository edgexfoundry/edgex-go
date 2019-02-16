#
# Copyright (c) 2018
# Dell
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
RUN make cmd/support-scheduler/support-scheduler

FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Dell, Cavium'

ENV APP_PORT=48085
#expose support scheduler port
EXPOSE $APP_PORT


COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/support-scheduler/support-scheduler /
COPY --from=builder /go/src/github.com/edgexfoundry/edgex-go/cmd/support-scheduler/res/docker/configuration.toml /res/docker/configuration.toml
ENTRYPOINT ["/support-scheduler","--consul","--profile=docker","--confdir=/res"]