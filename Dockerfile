# Docker image for Golang Core Data micro service 

# Initial Image - for building Golang
FROM golang:1.9.1-alpine AS build-env

RUN apk update
RUN apk add zeromq-dev
RUN apk add libsodium-dev
RUN apk add pkgconfig
# GCC compiler and build-tools
RUN apk add build-base

RUN mkdir -p /go/src \
 && mkdir -p /go/bin \
 && mkdir -p /go/pkg

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

# RUN mkdir -p $GOPATH/src/bitbucket.org/clientcto/go-core-data
WORKDIR src/github.com/edgexfoundry/core-data-go
COPY . .

RUN go build --ldflags '-extldflags "-lstdc++ -static -lsodium -static -lzmq"'

#Next image - Copy built Go binary into new workspace
FROM alpine:3.4

# Environment variables
ENV APP_DIR=/core-data-go
ENV APP_PORT=48080

# Expose data port
EXPOSE $APP_PORT

WORKDIR $APP_DIR
COPY --from=build-env /go/src/github.com/edgexfoundry/core-data-go/core-data-go .
COPY --from=build-env /go/src/github.com/edgexfoundry/core-data-go/docker-files ./res

ENTRYPOINT ./core-data-go
