# Notes For Adding A New Broker Type To Message Bus 

In order to add a new message bus broker type create a new folder and implement specific handler and add to messagebus_factory.go similar to the following code snippet
```go
case Mosquitto:
    mosquitto.Configure(ctx, cancel, f)
    return nil
```

Add a new compose file snippet for new broker type similar to the following example for Mosquitto Broker
```yml
version: '3.7'

volumes:
  mqtt:

services:
  mqtt-broker:
    image: eclipse-mosquitto:${MOSQUITTO_VERSION}
    entrypoint: ["/edgex-init/messagebus_wait_install.sh"]
    env_file:
      - common-security.env
      - common-sec-stage-gate.env
    environment:
      BROKER_TYPE: mosquitto
      CONF_DIR: /edgex-init/bootstrap-mqtt/res
      ENTRYPOINT_ARG: /usr/sbin/mosquitto -c /mosquitto/config/mosquitto.conf
    ports:
      - "127.0.0.1:1883:1883"
    volumes:
    - mqtt:/mosquitto:z
    - edgex-init:/edgex-init:ro,z
    - /tmp/edgex/secrets/security-bootstrapper-messagebus:/tmp/edgex/secrets/security-bootstrapper-messagebus:ro,z
    depends_on:
      - security-bootstrapper
      - secretstore-setup      
    container_name: edgex-mqtt-broker
    hostname: edgex-mqtt-broker
    read_only: true
    restart: always
    networks:
      - edgex-network
    security_opt:
      - no-new-privileges:true
    # root privilege required for bootstrapper's process
    user: root:root
```