FROM golang:1.12-alpine AS builder

ENV GO111MODULE=on

WORKDIR /edgex-go

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make git

COPY go.mod .

RUN go mod download

COPY . .

RUN make cmd/security-proxy-setup/security-proxy-setup

FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.'

WORKDIR /edgex

COPY --from=builder /edgex-go/cmd/security-proxy-setup/res/docker/configuration.toml res/configuration.toml

COPY --from=builder /edgex-go/cmd/security-proxy-setup/security-proxy-setup .

ENTRYPOINT ["/edgex/security-proxy-setup"]
CMD ["--init=true"]
