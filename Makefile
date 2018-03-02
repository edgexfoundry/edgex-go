#
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#

BUILD_DIR = $(PWD)/build

.PHONY: build clean test docker
GO = CGO_ENABLED=0 go
GOCGO = CGO_ENABLED=1 go

DOCKERS =
.PHONY: $(DOCKERS)

MICROSERVICES = export-client export-distro core-metadata core-data core-command
.PHONY: $(MICROSERVICES)

VERSION = $(shell cat ./VERSION)

GOFLAGS = -ldflags "-X github.com/edgexfoundry/edgex-go.Version=$(VERSION)"

build: $(MICROSERVICES)

core-metadata:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$@ ./cmd/$@ 

core-data:
	$(GOCGO) build $(GOFLAGS) -o $(BUILD_DIR)/$@ ./cmd/$@

core-command:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$@ ./cmd/$@

export-client:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$@ ./cmd/$@

export-distro:
	$(GOCGO) build $(GOFLAGS) -o $(BUILD_DIR)/$@ ./cmd/$@

clean:
	rm -rf $(BUILD_DIR)

test:
	go test `glide novendor`

prepare:
	glide install


docker: $(DOCKERS)

docker_core_metadata:
	docker build -f docker/Dockerfile.metadata -t edgexfoundry/docker-core-metadata .

docker_core_data:
	docker build -f docker/Dockerfile.data -t edgexfoundry/docker-core-data .

docker_core_command:
	docker build -f docker/Dockerfile.command -t edgexfoundry/docker-core-command .

docker_export_client:
	docker build -f docker/Dockerfile.client -t edgexfoundry/docker-export-client .

docker_export_distro:
	docker build -f docker/Dockerfile.distro -t edgexfoundry/docker-export-distro .
