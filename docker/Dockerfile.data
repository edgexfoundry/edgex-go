# Docker image for Golang Core Data micro service 

# Initial Image - for building Golang
FROM golang:1.9.1-alpine AS build-env

RUN apk update
# GCC compiler and build-tools
RUN apk add zeromq-dev libsodium-dev pkgconfig build-base git glide

RUN mkdir -p /go/src \
 && mkdir -p /go/bin \
 && mkdir -p /go/pkg

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

ENV CORE_DATA_GO=core-data-go
ENV CORE_DATA_GOPATH=src/github.com/edgexfoundry/$CORE_DATA_GO
ENV GO_CORE_DATA_REPO=https://github.com/edgexfoundry/core-data-go.git
RUN mkdir -p $GOPATH/$CORE_DATA_GOPATH
WORKDIR $CORE_DATA_GOPATH

COPY . .

RUN glide install \
 && go build --ldflags '-extldflags "-lstdc++ -static -lsodium -static -lzmq"'

ARG GOOS=linux
ARG GOARCH=amd64
ARG DATA_EXE=$CORE_DATA_GO

RUN go get -d
RUN GOOS=$GOOS GOARCH=$GOARCH go build --ldflags '-extldflags "-lstdc++ -static -lsodium -static -lzmq"' -o $DATA_EXE

#Next image - Copy built Go binary into new workspace
FROM alpine:3.6

# Environment variables
ENV GOPATH=/go
ENV CORE_DATA_GO=core-data-go
ENV CORE_DATA_PATH=core-data-go
ENV CORE_DATA_GOPATH=src/github.com/edgexfoundry/$CORE_DATA_GO
ENV APP_DIR=/$CORE_DATA_GO
ENV APP_PORT=48080

# Expose data port
EXPOSE $APP_PORT

WORKDIR $APP_DIR
COPY --from=build-env $GOPATH/$CORE_DATA_GOPATH/$CORE_DATA_GO .
COPY res/configuration-docker.json ./res/configuration.json

ENTRYPOINT ./$CORE_DATA_GO
