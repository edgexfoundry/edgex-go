# EdgeX Foundry Export Services
[![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/export-go)](https://goreportcard.com/report/github.com/edgexfoundry/export-go)
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX Export services.

[export-client](https://github.com/edgexfoundry/export-client),
[export-distro](https://github.com/edgexfoundry/export-distro) and
[export-domain](https://github.com/edgexfoundry/export-domain) have been merged
into one single repo. Additionaly, some classes from
[core-domain](https://github.com/edgexfoundry/core-domain) have been duplicated
into this repo (needed for linking and data isolation of microservices).

Repo contains two microservices, `export-client` and `export-distro`.

## Install and Deploy

To fetch the code and start the microservice execute:

```
go get github.com/edgexfoundry/export-go
cd $GOPATH/src/github.com/edgexfoundry/export-go
glide install
go run cmd/client/main.go
```
## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)
