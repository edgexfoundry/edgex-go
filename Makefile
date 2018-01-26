.PHONY: build test 

build:
	go build ./cmd/export-client
	go build ./cmd/export-distro

test:
	go test `glide novendor`

