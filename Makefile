#
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

.PHONY: build test docker

EXPORT_CLIENT_VERSION=$(shell cat export/client/VERSION)
EXPORT_DISTRO_VERSION=$(shell cat export/distro/VERSION)

GOFLAGS=-ldflags "-X client.version=$(EXPORT_CLIENT_VERSION) -X distro.version=$(EXPORT_DISTRO_VERSION)"

build:
	go build $(GOFLAGS) ./cmd/export-client
	go build $(GOFLAGS) ./cmd/export-distro

test:
	go test `glide novendor`

prepare:
	glide install

docker:
