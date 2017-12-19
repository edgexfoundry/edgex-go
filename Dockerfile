# Docker image for Golang Core Command micro service

# Initial Image - for building Golang
FROM golang:1.7.5-alpine AS build-env

# Dependencies
RUN apk add --update --no-cache \
    curl git && curl https://glide.sh/get | sh

# Setup Go env
RUN mkdir -p /go/src \
 && mkdir -p /go/bin \
 && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

ENV CORE_COMMAND_GO=core-command-go
ENV CORE_COMMAND_GOPATH=src/github.com/edgexfoundry/$CORE_COMMAND_GO
ENV GO_CORE_COMMAND_REPO=https://github.com/edgexfoundry/core-command-go.git
RUN mkdir -p $GOPATH/$CORE_COMMAND_GOPATH
WORKDIR $CORE_COMMAND_GOPATH

# Clone repo and install dependencies
COPY . .
RUN glide install \
 && go build --ldflags '-extldflags "-lstdc++ -static -lsodium -static -lzmq"'

ARG GOOS=linux
ARG GOARCH=amd64
ARG COMMAND_EXE=$CORE_COMMAND_GO

RUN go get -d
RUN GOOS=$GOOS GOARCH=$GOARCH go build -o $COMMAND_EXE

# Next image - Copy built Go binary into new workspace
FROM alpine:3.6

# environment variable
ENV GOPATH=/go
ENV CORE_COMMAND_GO=core-command-go
ENV CORE_COMMAND_GOPATH=src/github.com/edgexfoundry/$CORE_COMMAND_GO
ENV APP_DIR=/$CORE_COMMAND_GO
ENV APP_PORT=48082

#expose command port
EXPOSE $APP_PORT

WORKDIR $APP_DIR
COPY --from=build-env $GOPATH/$CORE_COMMAND_GOPATH/$CORE_COMMAND_GO .
COPY $GOPATH/$CORE_COMMAND_GOPATH/res/configuration-docker.json ./res/configuration.json

ENTRYPOINT ./$CORE_COMMAND_GO
