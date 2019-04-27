#
# Copyright (c) 2019
# Cavium
# Mainflux
#
# SPDX-License-Identifier: Apache-2.0
#

GIT_SHA=$(shell git rev-parse HEAD)
VERSION=$(shell cat ./VERSION)

BUILD_DIR := build

###
# For ARM cross-compilation add something like:
#
# CC = arm-linux-gnueabi-gcc
# CGO_LDFLAGS = -L/<path_to_libzmq>/src/.libs
# CGO_CFLAGS = -I/<path_to_libzmq>/include
# GOOS = linux
# GOARCH = arm
# GOARM = 7
###
FLAGS = GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} GO111MODULE=on
GO_FLAGS =  $(FLAGS) CGO_ENABLED=0 
CGO_FLAGS =  $(FLAGS) CC=${CC} CGO_LDFLAGS=${CGO_LDFLAGS} CGO_CFLAGS=${CGO_CFLAGS} CGO_ENABLED=1

SERVICES_GO := config-seed export-client core-metadata core-command support-logging support-notifications sys-mgmt-executor sys-mgmt-agent support-scheduler
SERVICES_CGO := core-data export-distro
SERVICES := $(SERVICES_GO) $(SERVICES_CGO)
DOCKERS = $(addprefix docker_,$(SERVICES))

.PHONY: $(SERVICES) $(DOCKERS) build clean test docker run

define compile_service
	$(1) go build -ldflags "-s -w -X github.com/edgexfoundry/edgex-go.Version=$(VERSION)" -o cmd/$(2)/$(2) cmd/$(2)/main.go
endef

define clean_services
	for svc in $(SERVICES); do \
		rm -rf cmd/$$svc/$$svc; \
	done
endef

define make_docker
	docker build \
		--no-cache \
		--label "git_sha=$(GIT_SHA)" \
		--build-arg SVC_NAME=$(subst docker_,,$(1)) \
		-t edgex/$(subst docker_,,$(1)):$(GIT_SHA) \
		-t edgex/$(subst docker_,,$(1)):$(VERSION)-dev \
		-f cmd/$(subst docker_,,$(1))/Dockerfile \
		.
endef

all: $(SERVICES)

$(SERVICES_GO):
	$(call compile_service,$(GO_FLAGS),$(@))
$(SERVICES_CGO):
	$(call compile_service,$(CGO_FLAGS),$(@))

clean:
	$(call clean_services)

test:
	GO111MODULE=on go test -cover ./...
	GO111MODULE=on go vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

run:
	cd scripts && ./run.sh

run_docker:
	cd scripts && ./run-docker.sh


$(DOCKERS):
	$(call make_docker,$(@))

dockers: $(DOCKERS)
