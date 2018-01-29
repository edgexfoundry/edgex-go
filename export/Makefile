# Copyright 2017 Cavium
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: build test vet prepare edgexclient edgexdistro edgexdistro_zmq docker

# Make exec targets phony to not track changes in go files. Compilation is fast
.PHONY: client distro distro_zmq

default: build

client:
	go build -o client cmd/client/main.go

distro:
	go build -o distro cmd/distro/main.go

distro_zmq:
	go build -o distro_zmq -tags zeromq cmd/distro/main.go

build: client distro distro_zmq

docker:
	docker build -f Dockerfile.client  .
	docker build -f Dockerfile.distro  .

test:
	go test `glide novendor`

vet:
	go vet `glide novendor` 

coverage:
	go test -covermode=count -coverprofile=cov.out ./distro
	go tool cover -html=cov.out -o distroCoverage.html
	go test -covermode=count -coverprofile=cov.out ./client
	go tool cover -html=cov.out -o clientCoverage.html
	rm cov.out

bench:
	go test -run=XXX -bench=. ./distro

profile:
	go test -run=XXX -bench=.  -cpuprofile distro.cpu ./distro
	go test -run=XXX -bench=.  -memprofile distro.mem ./distro

prepare:
	glide install

clean:
	rm -f client distro distro_zmq cov.out distroCoverage.html \
       clientCoverage.html distro.cpu distro.mem
