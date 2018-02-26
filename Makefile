.PHONY: build test docker

VERSION=$(shell cat VERSION)

GOFLAGS=-ldflags "-X client.version=$(VERSION) -X distro.version=$(VERSION)"

build:
	go build $(GOFLAGS) ./cmd/export-client
	go build $(GOFLAGS) ./cmd/export-distro

test:
	go test `glide novendor`

prepare:
	glide install

docker:
