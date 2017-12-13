# Docker image for Golang Core Meta Data micro service

# Initial Image - for building Golang
FROM golang:1.7.5-alpine AS build-env

# Dependencies
RUN apt-get update && apt-get install \
    curl && curl https://glide.sh/get | sh

# Setup Go env
RUN mkdir -p /go/src \
 && mkdir -p /go/bin \
 && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

ENV CORE_METADATA_GO=core-metadata-go
ENV CORE_METADATA_GOPATH=src/github.com/edgexfoundry/$CORE_METADATA_GO
ENV GO_CORE_METADATA_REPO=https://github.com/edgexfoundry/core-metadata-go.git
RUN mkdir -p $GOPATH/$CORE_METADATA_GOPATH
WORKDIR $CORE_METADATA_GOPATH

# Clone repo and install dependencies
RUN git clone $GO_CORE_METADATA_REPO . && glide install

ARG GOOS=linux
ARG GOARCH=amd64
ARG METADATA_EXE=$CORE_METADATA_GO

RUN GOOS=$GOOS GOARCH=$GOARCH go build -o $METADATA_EXE

# Next image - Copy built Go binary into new workspace
FROM alpine:3.6

# environment variable
ENV CORE_METADATA_GO=core-metadata-go
ENV CORE_METADATA_GOPATH=src/github.com/edgexfoundry/$CORE_METADATA_GO
ENV APP_DIR=/$CORE_METADATA_GO
ENV APP_PORT=48082

#expose meta data port
EXPOSE $APP_PORT

WORKDIR $APP_DIR
COPY --from=build-env $GOROOT/$CORE_METADATA_GOPATH/$CORE_METADATA_GO .
COPY --from=build-env $GOROOT/$CORE_METADATA_GOPATH/res ./res

ENTRYPOINT ./$CORE_METADATA_GO
