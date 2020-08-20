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

RUN apk update && apk add make bash git ca-certificates

COPY go.mod .

RUN go mod download

COPY . .
RUN make cmd/support-notifications/support-notifications

FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Cavium'

ENV APP_PORT=48060
#expose support notifications port
EXPOSE $APP_PORT

COPY --from=builder /bin/sh /bin/sh
COPY --from=builder /usr/share/ca-certificates /usr/share/ca-certificates
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /edgex-go/cmd/support-notifications/Attribution.txt /
COPY --from=builder /edgex-go/cmd/support-notifications/support-notifications /
COPY --from=builder /edgex-go/cmd/support-notifications/res/configuration.toml /res/configuration.toml
ENTRYPOINT ["/support-notifications"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]
