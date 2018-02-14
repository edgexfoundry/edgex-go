# Copyright 2017 Cavium
#
# SPDX-License-Identifier: Apache-2.0

BUILD_DIR := build

.PHONY: buildall test vet prepare $(BUILD_DIR)/client $(BUILD_DIR)/distro \
		$(BUILD_DIR)/distro_zmq docker

# Make exec targets phony to not track changes in go files. Compilation is fast
.PHONY: client distro distro_zmq

default: buildall

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

client: $(BUILD_DIR)
	go build -o $(BUILD_DIR)/client cmd/client/main.go

distro: $(BUILD_DIR)
	go build -o $(BUILD_DIR)/distro cmd/distro/main.go

distro_zmq: $(BUILD_DIR)
	go build -o $(BUILD_DIR)/distro_zmq -tags zeromq cmd/distro/main.go

buildall: client distro distro_zmq

docker:
	docker build -f Dockerfile.client  .
	docker build -f Dockerfile.distro  .

test:
	go test `glide novendor`

vet:
	go vet `glide novendor` 

coverage: $(BUILD_DIR)
	go test -covermode=count -coverprofile=$(BUILD_DIR)/cov.out ./distro
	go tool cover -html=$(BUILD_DIR)/cov.out -o $(BUILD_DIR)/distroCoverage.html
	go test -covermode=count -coverprofile=$(BUILD_DIR)/cov.out ./client
	go tool cover -html=$(BUILD_DIR)/cov.out -o $(BUILD_DIR)/clientCoverage.html
	rm $(BUILD_DIR)/cov.out

bench: $(BUILD_DIR)
	go test -run=XXX -bench=. ./distro

profile: $(BUILD_DIR)
	go test -run=XXX -bench=. -cpuprofile $(BUILD_DIR)/distro.cpu ./distro
	go test -run=XXX -bench=. -memprofile $(BUILD_DIR)/distro.mem ./distro

prepare:
	glide install

clean:
	rm -rf $(BUILD_DIR) distro.test
