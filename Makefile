#
# Copyright 2022 Intel Corporation
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#


.PHONY: build clean unittest hadolint lint test docker run

# change the following boolean flag to include or exclude the delayed start libs for builds for most of core services except support services
INCLUDE_DELAYED_START_BUILD_CORE:="false"
# change the following boolean flag to include or exclude the delayed start libs for builds for support services exculsively
INCLUDE_DELAYED_START_BUILD_SUPPORT:="true"

GO=CGO_ENABLED=0 GO111MODULE=on go

# see https://shibumi.dev/posts/hardening-executables
CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2"
CGO_CFLAGS="-O2 -pipe -fno-plt"
CGO_CXXFLAGS="-O2 -pipe -fno-plt"
CGO_LDFLAGS="-Wl,-O1,–sort-common,–as-needed,-z,relro,-z,now"
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
	docker_security_bootstrapper \
	docker_security_spire_server \
	docker_security_spire_agent \
	docker_security_spire_config \
	docker_security_spiffe_token_provider

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
	cmd/security-bootstrapper/security-bootstrapper \
	cmd/security-spiffe-token-provider/security-spiffe-token-provider

.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)
DOCKER_TAG=$(VERSION)-dev

GOFLAGS=-ldflags "-X github.com/edgexfoundry/edgex-go.Version=$(VERSION)" -trimpath -mod=readonly
CGOFLAGS=-ldflags "-linkmode=external -X github.com/edgexfoundry/edgex-go.Version=$(VERSION)" -trimpath -mod=readonly -buildmode=pie
GOTESTFLAGS?=-race

GIT_SHA=$(shell git rev-parse HEAD)

ARCH=$(shell uname -m)

GO_VERSION=$(shell grep '^go [0-9].[0-9]*' go.mod | cut -d' ' -f 2)

# DO NOT change the following flag, as it is automatically set based on the boolean switch INCLUDE_DELAYED_START_BUILD_CORE
NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE:=non_delayedstart
ifeq ($(INCLUDE_DELAYED_START_BUILD_CORE),"true")
	NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE:=
endif
NON_DELAYED_START_GO_BUILD_TAG_FOR_SUPPORT:=
ifeq ($(INCLUDE_DELAYED_START_BUILD_SUPPORT),"false")
	NON_DELAYED_START_GO_BUILD_TAG_FOR_SUPPORT:=non_delayedstart
endif

NO_MESSAGEBUS_GO_BUILD_TAG:=no_messagebus
NO_ZMQ_GO_BUILD_TAG:=no_zmq

# Base docker image to speed up local builds
BASE_DOCKERFILE=https://raw.githubusercontent.com/edgexfoundry/ci-build-images/golang-${GO_VERSION}/Dockerfile
LOCAL_CACHE_IMAGE_BASE=edgex-go-local-cache-base
LOCAL_CACHE_IMAGE=edgex-go-local-cache

build: $(MICROSERVICES)

build-nats:
	make -e ADD_BUILD_TAGS=include_nats_messaging build

tidy:
	go mod tidy

core: metadata data command

metadata: cmd/core-metadata/core-metadata
cmd/core-metadata/core-metadata:
	$(GO) build -tags "$(ADD_BUILD_TAGS) $(NO_ZMQ_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o $@ ./cmd/core-metadata

data: cmd/core-data/core-data
cmd/core-data/core-data:
	$(GOCGO) build -tags "$(ADD_BUILD_TAGS) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(CGOFLAGS) -o $@ ./cmd/core-data

command: cmd/core-command/core-command
cmd/core-command/core-command:
	$(GO) build -tags "$(ADD_BUILD_TAGS) $(NO_ZMQ_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o $@ ./cmd/core-command

support: notifications scheduler

notifications: cmd/support-notifications/support-notifications
cmd/support-notifications/support-notifications:
	$(GO) build -tags "$(ADD_BUILD_TAGS) $(NO_ZMQ_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_SUPPORT)" $(GOFLAGS) -o $@ ./cmd/support-notifications

scheduler: cmd/support-scheduler/support-scheduler
cmd/support-scheduler/support-scheduler:
	$(GO) build -tags "$(ADD_BUILD_TAGS) $(NO_ZMQ_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_SUPPORT)" $(GOFLAGS) -o $@ ./cmd/support-scheduler

cmd/sys-mgmt-executor/sys-mgmt-executor:
	$(GO) build -tags "$(NO_MESSAGEBUS_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o $@ ./cmd/sys-mgmt-executor

cmd/sys-mgmt-agent/sys-mgmt-agent:
	$(GO) build -tags "$(NO_MESSAGEBUS_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o $@ ./cmd/sys-mgmt-agent

proxy: cmd/security-proxy-setup/security-proxy-setup
cmd/security-proxy-setup/security-proxy-setup:
	$(GO) build -tags "$(NO_MESSAGEBUS_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o ./cmd/security-proxy-setup/security-proxy-setup ./cmd/security-proxy-setup

secretstore: cmd/security-secretstore-setup/security-secretstore-setup
cmd/security-secretstore-setup/security-secretstore-setup:
	$(GO) build -tags "$(NO_MESSAGEBUS_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o ./cmd/security-secretstore-setup/security-secretstore-setup ./cmd/security-secretstore-setup

token: cmd/security-file-token-provider/security-file-token-provider
cmd/security-file-token-provider/security-file-token-provider:
	$(GO) build -tags "$(NO_MESSAGEBUS_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o ./cmd/security-file-token-provider/security-file-token-provider ./cmd/security-file-token-provider

secrets-config: cmd/secrets-config/secrets-config
cmd/secrets-config/secrets-config:
	$(GO) build -tags "$(NO_MESSAGEBUS_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o ./cmd/secrets-config ./cmd/secrets-config

bootstrapper: cmd/security-bootstrapper/security-bootstrapper
cmd/security-bootstrapper/security-bootstrapper:
	$(GO) build -tags "$(NO_MESSAGEBUS_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o ./cmd/security-bootstrapper/security-bootstrapper ./cmd/security-bootstrapper

spiffetp: cmd/security-spiffe-token-provider/security-spiffe-token-provider
cmd/security-spiffe-token-provider/security-spiffe-token-provider:
	$(GO) build -tags "$(NO_MESSAGEBUS_GO_BUILD_TAG) $(NON_DELAYED_START_GO_BUILD_TAG_FOR_CORE)" $(GOFLAGS) -o $@ ./cmd/security-spiffe-token-provider

clean:
	rm -f $(MICROSERVICES)

unittest:
	GO111MODULE=on go test $(GOTESTFLAGS) -coverprofile=coverage.out ./...

hadolint:
	if which hadolint > /dev/null ; then hadolint --config .hadolint.yml `find * -type f -name 'Dockerfile*' -print` ; elif test "${ARCH}" = "x86_64" && which docker > /dev/null ; then docker run --rm -v `pwd`:/host:ro,z --entrypoint /bin/hadolint hadolint/hadolint:latest --config /host/.hadolint.yml `find * -type f -name 'Dockerfile*' | xargs -i echo '/host/{}'` ; fi
	
lint:
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run\n  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.46.2"
	@if [ "z${ARCH}" = "zx86_64" ] && which golangci-lint >/dev/null ; then golangci-lint run --config .golangci.yml ; else echo "WARNING: Linting skipped (not on x86_64 or linter not installed)"; fi


test: unittest hadolint lint
	GO111MODULE=on go vet ./...
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]
	./bin/test-attribution-txt.sh

docker: $(DOCKERS)

docker-nats:
	make -e ADD_BUILD_TAGS=include_nats_messaging docker

clean_docker_base:
	docker rmi -f $(LOCAL_CACHE_IMAGE) $(LOCAL_CACHE_IMAGE_BASE) 

docker_base:
	echo "Building local cache image";\
	response=$(shell curl --write-out '%{http_code}' --silent --output /dev/null "$(BASE_DOCKERFILE)"); \
	if [ "$${response}" = "200" ]; then \
		echo "Found base Dockerfile"; \
		curl -s "$(BASE_DOCKERFILE)" | docker build -t $(LOCAL_CACHE_IMAGE_BASE) -f - .; \
		echo "FROM $(LOCAL_CACHE_IMAGE_BASE)\nWORKDIR /edgex-go\nCOPY go.mod .\nRUN go mod download" | docker build -t $(LOCAL_CACHE_IMAGE) -f - .; \
	else \
		echo "No base Dockerfile found. Using golang:$(GO_VERSION)-alpine"; \
		echo "FROM golang:$(GO_VERSION)-alpine\nRUN apk add --update make git\nWORKDIR /edgex-go\nCOPY go.mod .\nRUN go mod download" | docker build -t $(LOCAL_CACHE_IMAGE) -f - .; \
	fi

dcore: dmetadata ddata dcommand

dmetadata: docker_core_metadata
docker_core_metadata: docker_base
	docker build \
		--build-arg ADD_BUILD_TAGS=$(ADD_BUILD_TAGS) \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/core-metadata/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/core-metadata:$(GIT_SHA) \
		-t edgexfoundry/core-metadata:$(DOCKER_TAG) \
		.

ddata: docker_core_data
docker_core_data: docker_base
	docker build \
		--build-arg ADD_BUILD_TAGS=$(ADD_BUILD_TAGS) \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/core-data/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/core-data:$(GIT_SHA) \
		-t edgexfoundry/core-data:$(DOCKER_TAG) \
		.

dcommand: docker_core_command
docker_core_command: docker_base
	docker build \
		--build-arg ADD_BUILD_TAGS=$(ADD_BUILD_TAGS) \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/core-command/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/core-command:$(GIT_SHA) \
		-t edgexfoundry/core-command:$(DOCKER_TAG) \
		.

dsupport: dnotifications dscheduler

dnotifications: docker_support_notifications
docker_support_notifications: docker_base
	docker build \
		--build-arg ADD_BUILD_TAGS=$(ADD_BUILD_TAGS) \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/support-notifications/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/support-notifications:$(GIT_SHA) \
		-t edgexfoundry/support-notifications:$(DOCKER_TAG) \
		.

dscheduler: docker_support_scheduler
docker_support_scheduler: docker_base
	docker build \
		--build-arg ADD_BUILD_TAGS=$(ADD_BUILD_TAGS) \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/support-scheduler/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/support-scheduler:$(GIT_SHA) \
		-t edgexfoundry/support-scheduler:$(DOCKER_TAG) \
		.

docker_sys_mgmt_agent: docker_base
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/sys-mgmt-agent/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/sys-mgmt-agent:$(GIT_SHA) \
		-t edgexfoundry/sys-mgmt-agent:$(DOCKER_TAG) \
		.

dproxy: docker_security_proxy_setup
docker_security_proxy_setup: docker_base
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/security-proxy-setup/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/security-proxy-setup:$(GIT_SHA) \
		-t edgexfoundry/security-proxy-setup:$(DOCKER_TAG) \
		.

dsecretstore: docker_security_secretstore_setup
docker_security_secretstore_setup: docker_base
		docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/security-secretstore-setup/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/security-secretstore-setup:$(GIT_SHA) \
		-t edgexfoundry/security-secretstore-setup:$(DOCKER_TAG) \
		.

dbootstrapper: docker_security_bootstrapper
docker_security_bootstrapper: docker_base
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/security-bootstrapper/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/security-bootstrapper:$(GIT_SHA) \
		-t edgexfoundry/security-bootstrapper:$(DOCKER_TAG) \
		.

dspires: docker_security_spire_server
docker_security_spire_server: docker_base
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/security-spire-server/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/security-spire-server:$(GIT_SHA) \
		-t edgexfoundry/security-spire-server:$(DOCKER_TAG) \
		.

dspirea: docker_security_spire_agent
docker_security_spire_agent: docker_base
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/security-spire-agent/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/security-spire-agent:$(GIT_SHA) \
		-t edgexfoundry/security-spire-agent:$(DOCKER_TAG) \
		.

dspirec: docker_security_spire_config
docker_security_spire_config: docker_base
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/security-spire-config/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/security-spire-config:$(GIT_SHA) \
		-t edgexfoundry/security-spire-config:$(DOCKER_TAG) \
		.

dspiffetp: docker_security_spiffe_token_provider
docker_security_spiffe_token_provider: docker_base
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg BUILDER_BASE=$(LOCAL_CACHE_IMAGE) \
		-f cmd/security-spiffe-token-provider/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/security-spiffe-token-provider:$(GIT_SHA) \
		-t edgexfoundry/security-spiffe-token-provider:$(DOCKER_TAG) \
		.

vendor:
	$(GO) mod vendor
