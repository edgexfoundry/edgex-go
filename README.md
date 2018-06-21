# EdgeX Foundry Go Services
[![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/edgex-go)](https://goreportcard.com/report/github.com/edgexfoundry/edgex-go) [![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX services.

All edgex go [repos](https://github.com/edgexfoundry/) have been merged into this repo.

The script [merge-edgex-go.sh](https://gist.github.com/feclare/8dba191e8cf77864fe5eed38b380f13a) has been used to generate this repo.

Specifically, launching the beginnings of this Git branch (whose purpose it is to bring in the support-notifications Go package into the EdgeX monorepo) with this preliminary update (to this README.md file).
## Install and Deploy

EdgeX Go code depends on ZMQ library. Make sure that you have dev version of the library
installed on your host.

For example, in the case of Debian Linux system you can:
```
sudo apt-get install libzmq3-dev
```

To fetch the code and compile the microservices execute:

```
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
glide install
make build
```
## Snap Package
Edgex Foundry is also available as a snap package, for more details
on the snap, including how to install it, please refer to [EdgeX snap](https://github.com/edgexfoundry/edgex-go/snap/README.md)


## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)
