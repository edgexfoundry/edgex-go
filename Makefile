#
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#


.PHONY: build clean test docker run

GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=CGO_ENABLED=1 GO111MODULE=on go

DOCKERS= \
	docker_core_data \
	docker_core_metadata \
	docker_core_command  \
	docker_support_notifications \
	docker_sys_mgmt_agent \
	docker_support_scheduler \
	docker_security_proxy_setup \
	docker_security_secretstore_setup \
	docker_security_bootstrapper

.PHONY: $(DOCKERS)

MICROSERVICES= \
	cmd/core-data/core-data \
	cmd/core-metadata/core-metadata \
	cmd/core-command/core-command \
	cmd/support-notifications/support-notifications \
	cmd/sys-mgmt-executor/sys-mgmt-executor \
	cmd/sys-mgmt-agent/sys-mgmt-agent \
	cmd/support-scheduler/support-scheduler \
	cmd/security-proxy-setup/security-proxy-setup \
	cmd/security-secretstore-setup/security-secretstore-setup \
	cmd/security-file-token-provider/security-file-token-provider \
	cmd/secrets-config/secrets-config \
	cmd/security-bootstrapper/security-bootstrapper

.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)
DOCKER_TAG=$(VERSION)-dev

GOFLAGS=-ldflags "-X github.com/edgexfoundry/edgex-go.Version=$(VERSION)"
GOTESTFLAGS?=-race

GIT_SHA=$(shell git rev-parse HEAD)

ARCH=$(shell uname -m)

build: $(MICROSERVICES)

cmd/core-metadata/core-metadata:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/core-metadata

cmd/core-data/core-data:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/core-data

cmd/core-command/core-command:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/core-command

cmd/support-notifications/support-notifications:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/support-notifications

cmd/sys-mgmt-executor/sys-mgmt-executor:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/sys-mgmt-executor

cmd/sys-mgmt-agent/sys-mgmt-agent:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/sys-mgmt-agent

cmd/support-scheduler/support-scheduler:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/support-scheduler

cmd/security-proxy-setup/security-proxy-setup:
	$(GOCGO) build $(GOFLAGS) -o ./cmd/security-proxy-setup/security-proxy-setup ./cmd/security-proxy-setup

cmd/security-secretstore-setup/security-secretstore-setup:
	$(GOCGO) build $(GOFLAGS) -o ./cmd/security-secretstore-setup/security-secretstore-setup ./cmd/security-secretstore-setup

cmd/security-file-token-provider/security-file-token-provider:
	$(GOCGO) build $(GOFLAGS) -o ./cmd/security-file-token-provider/security-file-token-provider ./cmd/security-file-token-provider

cmd/secrets-config/secrets-config:
	$(GOCGO) build $(GOFLAGS) -o ./cmd/secrets-config ./cmd/secrets-config

cmd/security-bootstrapper/security-bootstrapper:
	$(GOCGO) build $(GOFLAGS) -o ./cmd/security-bootstrapper/security-bootstrapper ./cmd/security-bootstrapper

clean:
	rm -f $(MICROSERVICES)

test:
	GO111MODULE=on go test $(GOTESTFLAGS) -coverprofile=coverage.out ./...
	GO111MODULE=on go vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]
	./bin/test-go-mod-tidy.sh
	./bin/test-attribution-txt.sh

run:
	cd bin && ./edgex-launch.sh

run_docker:
	bin/edgex-docker-launch.sh $(EDGEX_DB)

docker: $(DOCKERS)

docker_core_metadata:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/core-metadata/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-core-metadata-go:$(GIT_SHA) \
		-t edgexfoundry/docker-core-metadata-go:$(DOCKER_TAG) \
		.

docker_core_data:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/core-data/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-core-data-go:$(GIT_SHA) \
		-t edgexfoundry/docker-core-data-go:$(DOCKER_TAG) \
		.

docker_core_command:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/core-command/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-core-command-go:$(GIT_SHA) \
		-t edgexfoundry/docker-core-command-go:$(DOCKER_TAG) \
		.

docker_support_notifications:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/support-notifications/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-support-notifications-go:$(GIT_SHA) \
		-t edgexfoundry/docker-support-notifications-go:$(DOCKER_TAG) \
		.

docker_support_scheduler:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/support-scheduler/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-support-scheduler-go:$(GIT_SHA) \
		-t edgexfoundry/docker-support-scheduler-go:$(DOCKER_TAG) \
		.

docker_sys_mgmt_agent:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/sys-mgmt-agent/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-sys-mgmt-agent-go:$(GIT_SHA) \
		-t edgexfoundry/docker-sys-mgmt-agent-go:$(DOCKER_TAG) \
		.

docker_security_proxy_setup:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/security-proxy-setup/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-security-proxy-setup-go:$(GIT_SHA) \
		-t edgexfoundry/docker-security-proxy-setup-go:$(DOCKER_TAG) \
		.

docker_security_secretstore_setup:
		docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/security-secretstore-setup/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-security-secretstore-setup-go:$(GIT_SHA) \
		-t edgexfoundry/docker-security-secretstore-setup-go:$(DOCKER_TAG) \
		.

docker_security_bootstrapper:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/security-bootstrapper/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-security-bootstrapper-go:$(GIT_SHA) \
		-t edgexfoundry/docker-security-bootstrapper-go:$(DOCKER_TAG) \
		.
