# EdgeX Foundry Go Services
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX services.

All edgex go [repos](https://github.com/edgexfoundry/) have been merged into this repo.

The script [merge-edgex-go.sh](https://gist.github.com/feclare/8dba191e8cf77864fe5eed38b380f13a) has been used to generate this repo.
## Install and Deploy

Currently only `export-client` and `export-distro` are functional. `core-command`, `core-metadata` and `core-data` can be compiled but doesn't have a command in cmd

To fetch the code and compile the microservices execute:

```
go get github.com/feclare/edgex-go
cd /home/feclare/projects/gopath/src/github.com/feclare/edgex-go
glide install
make build
```
## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)
