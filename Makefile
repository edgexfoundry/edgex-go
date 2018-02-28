#
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

.PHONY: build test docker

EXPORT_CLIENT_VERSION=$(shell cat export/client/VERSION)
EXPORT_DISTRO_VERSION=$(shell cat export/distro/VERSION)
CORE_DATA_VERSION=$(shell cat core/data/VERSION)
CORE_METADATA_VERSION=$(shell cat core/metadata/VERSION)
CORE_COMMAND_VERSION=$(shell cat core/command/VERSION)

GOFLAGS=-ldflags "-X client.version=$(EXPORT_CLIENT_VERSION) -X distro.version=$(EXPORT_DISTRO_VERSION)"

build:
	go build $(GOFLAGS) ./cmd/export-client
	go build $(GOFLAGS) ./cmd/export-distro
	go build $(GOFLAGS) ./cmd/core-metadata
	go build $(GOFLAGS) ./cmd/core-command
	go build $(GOFLAGS) ./cmd/core-data

test:
	go test `glide novendor`

prepare:
	glide install

docker:
