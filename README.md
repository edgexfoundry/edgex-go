# EdgeX Foundry Go Services
[![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/edgex-go)](https://goreportcard.com/report/github.com/edgexfoundry/edgex-go)
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX services.

All edgex go [repos](https://github.com/edgexfoundry/) have been merged into this repo.

The script [merge-edgex-go.sh](https://gist.github.com/feclare/8dba191e8cf77864fe5eed38b380f13a) has been used to generate this repo.

## Install

### From Source
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

### Snap Package
EdgeX Foundry is also available as a snap package, for more details
on the snap, including how to install it, please refer to [EdgeX snap](https://github.com/edgexfoundry/edgex-go/snap/README.md)

## Deploy
EdgeX system can be deployed as a set of Docker containers or a previously compiler binaries.

### Docker
EdgeX images are kept on organization's [DockerHub page](https://hub.docker.com/u/edgexfoundry/).
They can be run in orchestration via official [docker-compose.yml](https://github.com/edgexfoundry/developer-scripts/blob/master/compose-files/docker-compose.yml).

Simplest way is to do this via prepared script in `bin` directory:
```
cd bin 
./edgex-docker-launch.sh
```

### Compiled Binaries
During development phase, it is important to run compiled binaries (not containers).

There is a script in `bin` directory that can help you launch the whole EdgeX system:
```
cd bin
./edgex-launch.sh
```

## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)
