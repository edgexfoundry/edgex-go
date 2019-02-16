#
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#


.PHONY: build clean test docker run


GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=CGO_ENABLED=1 GO111MODULE=on go

DOCKERS=docker_config_seed docker_export_client docker_export_distro docker_core_data docker_core_metadata docker_core_command docker_support_logging docker_support_notifications docker_sys_mgmt_agent docker_support_scheduler
.PHONY: $(DOCKERS)

MICROSERVICES=cmd/config-seed/config-seed cmd/export-client/export-client cmd/export-distro/export-distro cmd/core-metadata/core-metadata cmd/core-data/core-data cmd/core-command/core-command cmd/support-logging/support-logging cmd/support-notifications/support-notifications cmd/sys-mgmt-agent/sys-mgmt-agent cmd/support-scheduler/support-scheduler

.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/edgex-go.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)

cmd/config-seed/config-seed:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/config-seed

cmd/core-metadata/core-metadata:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/core-metadata

cmd/core-data/core-data:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/core-data

cmd/core-command/core-command:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/core-command

cmd/export-client/export-client:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/export-client

cmd/export-distro/export-distro:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/export-distro

cmd/support-logging/support-logging:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-logging

cmd/support-notifications/support-notifications:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-notifications

cmd/sys-mgmt-agent/sys-mgmt-agent:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/sys-mgmt-agent

cmd/support-scheduler/support-scheduler:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-scheduler

clean:
	rm -f $(MICROSERVICES)

test:
	GO111MODULE=on go test -cover ./...
	GO111MODULE=on go vet ./...

prepare:

run:
	cd bin && ./edgex-launch.sh

run_docker:
	cd bin && ./edgex-docker-launch.sh

docker: $(DOCKERS)

docker_config_seed:
	docker build \
		-f cmd/config-seed/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-core-config-seed-go:$(GIT_SHA) \
		-t edgexfoundry/docker-core-config-seed-go:$(VERSION)-dev \
		.

docker_core_metadata:
	docker build \
		-f cmd/core-metadata/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-core-metadata-go:$(GIT_SHA) \
		-t edgexfoundry/docker-core-metadata-go:$(VERSION)-dev \
		.

docker_core_data:
	docker build \
		-f cmd/core-data/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-core-data-go:$(GIT_SHA) \
		-t edgexfoundry/docker-core-data-go:$(VERSION)-dev \
		.

docker_core_command:
	docker build \
		-f cmd/core-command/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-core-command-go:$(GIT_SHA) \
		-t edgexfoundry/docker-core-command-go:$(VERSION)-dev \
		.

docker_export_client:
	docker build \
		-f cmd/export-client/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-export-client-go:$(GIT_SHA) \
		-t edgexfoundry/docker-export-client-go:$(VERSION)-dev \
		.

docker_export_distro:
	docker build \
		-f cmd/export-distro/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-export-distro-go:$(GIT_SHA) \
		-t edgexfoundry/docker-export-distro-go:$(VERSION)-dev \
		.

docker_support_logging:
	docker build \
		-f cmd/support-logging/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-support-logging-go:$(GIT_SHA) \
		-t edgexfoundry/docker-support-logging-go:$(VERSION)-dev \
		.

docker_support_notifications:
	docker build \
		-f cmd/support-notifications/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-support-notifications-go:$(GIT_SHA) \
		-t edgexfoundry/docker-support-notifications-go:$(VERSION)-dev \
		.

docker_support_scheduler:
	docker build \
		-f cmd/support-scheduler/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-support-scheduler-go:$(GIT_SHA) \
		-t edgexfoundry/docker-support-scheduler-go:$(VERSION)-dev \
		.
