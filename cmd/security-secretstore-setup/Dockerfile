FROM golang:1.13-alpine AS builder

ENV GO111MODULE=on

WORKDIR /edgex-go

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make git

COPY go.mod .

RUN go mod download

COPY . .

RUN make cmd/security-file-token-provider/security-file-token-provider cmd/security-secretstore-setup/security-secretstore-setup

FROM alpine:3.10

RUN apk update && apk add ca-certificates dumb-init curl \
    && rm -rf /var/cache/apk/* 

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.'

WORKDIR /

COPY --from=builder /edgex-go/cmd/security-file-token-provider/res/token-config.json /res-file-token-provider/token-config.json
COPY --from=builder /edgex-go/cmd/security-secretstore-setup/res-file-token-provider/docker/configuration.toml  /res-file-token-provider/configuration.toml
COPY --from=builder /edgex-go/cmd/security-secretstore-setup/res/docker/configuration.toml /res/configuration.toml

COPY --from=builder /edgex-go/cmd/security-file-token-provider/security-file-token-provider .
COPY --from=builder /edgex-go/cmd/security-secretstore-setup/security-secretstore-setup .

# setup the entry point script
COPY --from=builder /edgex-go/cmd/security-secretstore-setup/entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh \
    && ln -s /usr/local/bin/entrypoint.sh /

ENTRYPOINT ["entrypoint.sh"]
