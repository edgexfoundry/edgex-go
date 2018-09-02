#!/bin/bash
#
# Copyright (c) 2018
# Mainflux
#
# SPDX-License-Identifier: Apache-2.0
#

###
# Core Metadata
###
printf "> Provisioning Addressibles\n"

# Addressable for the Device Service
curl http://localhost:48081/api/v1/addressable -X POST -s -S -d @- <<EOF
{
    "name":"camera control",
    "protocol":"HTTP",
    "address":"172.17.0.1",
    "port":49977,
    "path":"/cameracontrol",
    "publisher":"none",
    "user":"none",
    "password":"none",
    "topic":"none"
}
EOF

echo ""

# Addressable for the Device (the camera in this case)
curl http://localhost:48081/api/v1/addressable -X POST -s -S -d @- <<EOF
{
    "name":"camera1 address",
    "protocol":"HTTP",
    "address":"172.17.0.1",
    "port":49999,
    "path":"/camera1",
    "publisher":"none",
    "user":"none",
    "password":"none",
    "topic":"none"
}
EOF

###
# Core Data
###
printf "\n> Provisioning Value Descriptors\n"

# Value Descriptor for `people count`
curl http://localhost:48080/api/v1/valuedescriptor -X POST -s -S -d @- <<EOF
{
    "name":"humancount",
    "description":"people count",
    "min":"0","max":"100",
    "type":"I",
    "uomLabel":"count",
    "defaultValue":"0",
    "formatting":"%s",
    "labels":["count","humans"]
}
EOF

echo ""

# Value Descriptor for `dog count`
curl http://localhost:48080/api/v1/valuedescriptor -X POST -s -S -d @- <<EOF
{
    "name":"caninecount",
    "description":"dog count",
    "min":"0","max":"100",
    "type":"I",
    "uomLabel":"count",
    "defaultValue":"0",
    "formatting":"%s",
    "labels":["count","canines"]
}
EOF

echo ""

# Value Descriptor for `camera scan distance`
curl http://localhost:48080/api/v1/valuedescriptor -X POST -s -S -d @- <<EOF
{
    "name":"depth",
    "description":"scan distance",
    "min":"1","max":"10",
    "type":"I",
    "uomLabel":"feet",
    "defaultValue":"1",
    "formatting":"%s",
    "labels":["scan","distance"]
}
EOF

echo ""

# Value Descriptor for `time between events`
curl http://localhost:48080/api/v1/valuedescriptor -X POST -s -S -d @- <<EOF
{
    "name":"duration",
    "description":"time between events",
    "min":"10",
    "max":"180",
    "type":"I",
    "uomLabel":"seconds",
    "defaultValue":"10",
    "formatting":"%s",
    "labels":["duration","time"]
}
EOF

echo ""

# Value Descriptor for `error`
curl http://localhost:48080/api/v1/valuedescriptor -X POST -s -S -d @- <<EOF
{
    "name":"cameraerror",
    "description":"error response message from a camera",
    "min":"",
    "max":"",
    "type":"S",
    "uomLabel":"",
    "defaultValue":"error",
    "formatting":"%s",
    "labels":["error","message"]
}
EOF

###
# Create Device Profile
###
printf "\n> Creating Device Profile\n"

# N.B camera_monitor_profile.yml file can be downloaded from https://docs.edgexfoundry.org/Ch-WalkthroughDeviceProfile.html 
curl http://localhost:48081/api/v1/deviceprofile/uploadfile -X POST -s -S -F "file=@./camera_monitor_profile.yml"

###
# Create Device Service
###
printf "\n> Creating Device Service\n"
curl http://localhost:48081/api/v1/deviceservice -X POST -s -S -d @- <<EOF
{
    "name":"camera control device service",
    "description":"Manage human and dog counting cameras",
    "labels":["camera","counter"],
    "adminState":"unlocked",
    "operatingState":"enabled",
    "addressable": {
        "name":"camera control"
    }
}
EOF

###
# Add Device
###
printf "\n> Creating Device\n"
curl http://localhost:48081/api/v1/device -X POST -s -S -d @- <<EOF
{
    "name":"countcamera1",
    "description":"human and dog counting camera #1",
    "adminState":"unlocked",
    "operatingState":"enabled",
    "addressable":{
        "name":"camera1 address"
    },
    "labels":[
        "camera","counter"
    ],
    "location":"",
    "service":{
        "name":"camera control device service"
    },
    "profile":{
        "name":"camera monitor profile"
    }
}
EOF

###
# Add Events
###
printf "\n> Creating Events\n"
curl http://localhost:48080/api/v1/event -X POST -s -S -d @- <<EOF
{
    "device":"countcamera1",
    "readings":[
        {
            "name":"humancount",
            "value":"5"
        },
        {
            "name":"caninecount",
            "value":"3"
        }
    ]
}
EOF

echo ""

curl http://localhost:48080/api/v1/event -X POST -s -S -d @- <<EOF
{
    "device":"countcamera1",
    "origin":1471806386919,
    "readings":[
        {
            "name":"humancount",
            "value":"1",
            "origin":1471806386919
        },
        {
            "name":"caninecount",
            "value":"0",
            "origin":1471806386919
        }
    ]
}
EOF

###
# Register Export Client
###
printf "\n> Registering Export Client\n"

curl http://localhost:48071/api/v1/registration -X POST -s -S -d @- <<EOF
{
    "name":"myMqttPublisher",
    "addressable":{
        "name":"myMqttBroker",
        "protocol":"tcp",
        "address":"test.mosquitto.org",
        "port":1883,
        "publisher":"EdgeXExportPublisher",
        "user":"",
        "password":"",
        "topic":"edgex/data"
    },
    "format":"JSON",
    "encryption":{
        "encryptionAlgorithm":"",
        "encryptionKey":"",
        "initializingVector":""
    },
    "enable":true,
    "destination":"MQTT_TOPIC"
}
EOF
