#
# Copyright (c) 2018
# Dell
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

RUN apk update && apk add make git

COPY go.mod .

RUN go mod download

COPY . .
RUN make cmd/support-scheduler/support-scheduler

FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Dell, Cavium'

ENV APP_PORT=48085
#expose support scheduler port
EXPOSE $APP_PORT

COPY --from=builder /edgex-go/cmd/support-scheduler/Attribution.txt /
COPY --from=builder /edgex-go/cmd/support-scheduler/support-scheduler /
COPY --from=builder /edgex-go/cmd/support-scheduler/res/configuration.toml /res/configuration.toml
ENTRYPOINT ["/support-scheduler","-cp=consul.http://edgex-core-consul:8500","--registry","--confdir=/res"]