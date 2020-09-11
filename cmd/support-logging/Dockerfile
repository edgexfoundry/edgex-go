#
# Copyright (c) 2018
# Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

ARG BUILDER_BASE=golang:1.15-alpine
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make && apk add bash git

COPY go.mod .

RUN go mod download

COPY . .
RUN make cmd/support-logging/support-logging

#Next image - Copy built Go binary into new workspace
FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Cavium'

ENV APP_PORT=48061
#expose support logging port
EXPOSE $APP_PORT

COPY --from=builder /bin/bash /bin/bash
COPY --from=builder /edgex-go/cmd/support-logging/Attribution.txt /
COPY --from=builder /edgex-go/cmd/support-logging/support-logging /
COPY --from=builder /edgex-go/cmd/support-logging/res/configuration.toml /res/configuration.toml
ENTRYPOINT ["/support-logging"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]
