# EdgeX Provision Script

This script facilitates example provisioning procedure by executing
HTTP requests as described in an official documentation
under [EdgeX Demonstration API Walk Through](https://docs.edgexfoundry.org/Ch-Walkthrough.html) chapter.

Additionaly, this procedure has been described in the blog post ["EdgeX in Anger, Part 2: Provisioning"](https://medium.com/mainflux-iot-platform/edgex-in-anger-part-2-provisioning-9d2da7a5cb27).

## Usage
Just execute the script from the current directory (so that it can find `camera_monitor_profile.yml` file):

```
./provision.sh
```

To see if Value Descriptors were configured properly, execute:
```
curl http://localhost:48080/api/v1/valuedescriptor | jsonpp
```

To check Device Service provisioning:
```
curl http://localhost:48081/api/v1/deviceservice | jsonpp
```

To check Device provisioning:
```
curl http://localhost:48081/api/v1/device
```

And so on. You will find more examples in the official [EdgeX documentation](https://docs.edgexfoundry.org/Ch-Walkthrough.html).
