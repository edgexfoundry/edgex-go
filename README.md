# go-core-metadata
includes the device/sensor metadata database and APIs to expose the database to other services. In particular, the device provisioning service will deposit and manage device metadata through this service. This service may also hold and manage other configuration metadata used by other services on the gateway – such as clean up schedules, hardware configuration (Wi-Fi connection info, MQTT queues, etc.). Non-device metadata may need to be held in a different database and/or managed by another service – depending on implementation

## TODO ## 
- add device profile by yaml
- break up into seperate packages
- Check for logging only at lowest possible level
- Check that String is being checked as bson.ObjectId at the mgo level
- ~~rest_devicereport~~
- ~~refactor device service~~
- ~~refactor schedule event~~


## TODO 3.0 (in order)
- ~~Dockerize~~
- Docker Compose file
- ProvisionWatchers (Callbacks)
- DB Reff (Doesn't work well with golang mgo package)

###Extra
- Consul
- (not a big deal)
- logging service
- notifications

# Docker
## Build
- `docker build -t go-core-metadata-docker .`
## Run
- `docker run -p 48081:48081 --name fuse-core-metadata --volumes-from fuse-files -d go-core-metadata-docker`

## BusyBox Build
- build your go container using the following commands `GOOS=linux GOARCH=386 go build .`
