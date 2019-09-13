FROM golang:1.12-alpine AS builder

ENV GO111MODULE=on

WORKDIR /edgex-go

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make git

COPY go.mod .

RUN go mod download

COPY . .

RUN make cmd/security-secretstore-setup/security-secretstore-setup

FROM alpine:3.10

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.'

WORKDIR /

COPY --from=builder /edgex-go/cmd/security-secretstore-setup/res/docker/configuration.toml /res/configuration.toml

COPY --from=builder /edgex-go/cmd/security-secretstore-setup/security-secretstore-setup .

ENTRYPOINT ["./security-secretstore-setup","--init=true"]
