GO=GO111MODULE=on go
HOOKS=cmd/configure/configure cmd/install/install

tidy:
	go mod tidy

build: tidy $(HOOKS)

cmd/configure/configure:
	$(GO) build -o $@ ./cmd/configure

cmd/install/install:
	$(GO) build -o $@ ./cmd/install

# TODO: fixup unit testing
test:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

clean:
	rm -f $(HOOKS)

