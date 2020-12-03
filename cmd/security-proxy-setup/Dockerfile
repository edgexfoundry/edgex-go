ARG BUILDER_BASE=golang:1.15-alpine
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make git

COPY go.mod .

RUN go mod download

COPY . .

RUN make cmd/security-proxy-setup/security-proxy-setup cmd/secrets-config/secrets-config

FROM alpine

RUN apk update && apk add curl && rm -rf /var/cache/apk/* 

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.'

WORKDIR /edgex

COPY --from=builder /edgex-go/cmd/security-proxy-setup/res/configuration.toml res/configuration.toml
COPY --from=builder /edgex-go/cmd/security-proxy-setup/security-proxy-setup .
# Note that secrets-config shares the same configuration file as security-proxy-setup
# as we are splitting security-proxy-setup into two different utilities for ease-of-use.
COPY --from=builder /edgex-go/cmd/secrets-config/secrets-config .

ENTRYPOINT ["/edgex/security-proxy-setup"]
CMD ["--init=true"]
