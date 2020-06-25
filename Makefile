#
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#


.PHONY: build clean test docker run

GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=CGO_ENABLED=1 GO111MODULE=on go

DOCKERS=docker_core_data docker_core_metadata docker_core_command docker_support_logging docker_support_notifications docker_sys_mgmt_agent docker_support_scheduler docker_security_secrets_setup docker_security_proxy_setup docker_security_secretstore_setup
.PHONY: $(DOCKERS)

MICROSERVICES=cmd/core-metadata/core-metadata cmd/core-data/core-data \
  cmd/core-command/core-command cmd/support-logging/support-logging \
	cmd/support-notifications/support-notifications cmd/sys-mgmt-executor/sys-mgmt-executor \
	cmd/sys-mgmt-agent/sys-mgmt-agent cmd/support-scheduler/support-scheduler \
	cmd/security-secrets-setup/security-secrets-setup cmd/security-proxy-setup/security-proxy-setup \
	cmd/security-secretstore-setup/security-secretstore-setup \
	cmd/security-file-token-provider/security-file-token-provider cmd/security-secretstore-read/security-secretstore-read

.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)
DOCKER_TAG=$(VERSION)-dev

GOFLAGS=-ldflags "-X github.com/edgexfoundry/edgex-go.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

ARCH=$(shell uname -m)

build: $(MICROSERVICES)

cmd/core-metadata/core-metadata:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/core-metadata

cmd/core-data/core-data:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/core-data

cmd/core-command/core-command:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/core-command

cmd/support-logging/support-logging:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-logging

cmd/support-notifications/support-notifications:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-notifications

cmd/sys-mgmt-executor/sys-mgmt-executor:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/sys-mgmt-executor

cmd/sys-mgmt-agent/sys-mgmt-agent:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/sys-mgmt-agent

cmd/support-scheduler/support-scheduler:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/support-scheduler

cmd/security-secrets-setup/security-secrets-setup:
	$(GO) build $(GOFLAGS) -o ./cmd/security-secrets-setup/security-secrets-setup ./cmd/security-secrets-setup

cmd/security-proxy-setup/security-proxy-setup:
	$(GO) build $(GOFLAGS) -o ./cmd/security-proxy-setup/security-proxy-setup ./cmd/security-proxy-setup

cmd/security-secretstore-setup/security-secretstore-setup:
	$(GO) build $(GOFLAGS) -o ./cmd/security-secretstore-setup/security-secretstore-setup ./cmd/security-secretstore-setup

cmd/security-file-token-provider/security-file-token-provider:
	$(GO) build $(GOFLAGS) -o ./cmd/security-file-token-provider/security-file-token-provider ./cmd/security-file-token-provider

cmd/security-secretstore-read/security-secretstore-read:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/security-secretstore-read

clean:
	rm -f $(MICROSERVICES)

test:
	GO111MODULE=on go test -race -coverprofile=coverage.out ./...
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

docker_support_logging:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/support-logging/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-support-logging-go:$(GIT_SHA) \
		-t edgexfoundry/docker-support-logging-go:$(DOCKER_TAG) \
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

docker_security_secrets_setup:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/security-secrets-setup/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-secrets-setup-go:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-secrets-setup-go:$(DOCKER_TAG) \
		.

docker_security_proxy_setup:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/security-proxy-setup/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-security-proxy-setup-go:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-security-proxy-setup-go:$(DOCKER_TAG) \
		.

docker_security_secretstore_setup:
		docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f cmd/security-secretstore-setup/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-edgex-security-secretstore-setup-go:$(GIT_SHA) \
		-t edgexfoundry/docker-edgex-security-secretstore-setup-go:$(DOCKER_TAG) \
		.

raml_verify:
	docker run --rm --privileged \
		-v $(PWD):/raml-verification -w /raml-verification \
		nexus3.edgexfoundry.org:10003/edgex-docs-builder:$(ARCH) \
		/scripts/raml-verify.sh
